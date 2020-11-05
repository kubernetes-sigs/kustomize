package builtins_qlik

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
	"github.com/sosedoff/gitkit"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

type testExecutableResolverT struct {
	path string
	err  error
}

func (r *testExecutableResolverT) Executable() (string, error) {
	return r.path, r.err
}

func Test_GoGetter(t *testing.T) {
	type tcT struct {
		name            string
		pluginConfig    string
		loaderRootDir   string
		teardown        func(*testing.T)
		checkAssertions func(*testing.T, resmap.ResMap)
	}
	testCases := []*tcT{
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			gitService := gitkit.New(gitkit.Config{
				Dir:        tmpDir,
				AutoCreate: true,
			})
			if err := gitService.Setup(); err != nil {
				t.Fatalf("error starting gitkit service: %v", err)
			}
			gitServer := httptest.NewServer(gitService)

			if _, err := execCmd(tmpDir, "git", "clone", fmt.Sprintf("%s/foo.git", gitServer.URL), "foo"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if _, err := execCmd(path.Join(tmpDir, "foo"), "git", "config", "user.email", "you@example.com"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if err := ioutil.WriteFile(path.Join(tmpDir, "foo", "kustomization.yaml"), []byte(`
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
- name: foo-config
  literals:
  - foo=bar
`), os.ModePerm); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if _, err := execCmd(path.Join(tmpDir, "foo"), "git", "add", "."); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if _, err := execCmd(path.Join(tmpDir, "foo"), "git", "commit", "-m", "First commit"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if _, err := execCmd(path.Join(tmpDir, "foo"), "git", "push", "origin", "master"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			return &tcT{
				name: "go-get and kustomize",
				pluginConfig: fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: GoGetter
metadata:
 name: notImportantHere
url: %s
preBuildScript: |-
    package main
    import (
        "fmt"
        "io/ioutil"
        yaml "gopkg.in/yaml.v3"
    )
        
    type Kustomization struct {
        GeneratorOptions struct {
            DisableNameSuffixHash bool `+"`yaml:\"disableNameSuffixHash\"`"+`
        } `+"`yaml:\"generatorOptions\"`"+`
        ConfigMapGenerator []struct {
            Name     string   `+"`yaml:\"name\"`"+`
            Literals []string `+"`yaml:\"literals\"`"+`
        } `+"`yaml:\"configMapGenerator\"`"+`
    }
    
    func main() {
        var k Kustomization
        yamlFile, err := ioutil.ReadFile("kustomization.yaml")
        if err != nil {
            panic(err)
        }
        err = yaml.Unmarshal(yamlFile, &k)
        if err != nil {
            panic(err)
        }
        k.ConfigMapGenerator[0].Literals[0] = "foo=changebar"
        b, err := yaml.Marshal(k)
        if err != nil {
            panic(err)
        }
        err = ioutil.WriteFile("kustomization.yaml", b, 0644)
        if err != nil {
            panic(err)
        }
    }
`, fmt.Sprintf("git::%s/foo", gitServer.URL)),
				loaderRootDir: tmpDir,
				teardown: func(t *testing.T) {
					gitServer.Close()
					_ = os.RemoveAll(tmpDir)
				},
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expectedK8sYaml := `apiVersion: v1
data:
  foo: changebar
kind: ConfigMap
metadata:
  name: foo-config
`
					if resMapYaml, err := resMap.AsYaml(); err != nil {
						t.Fatalf("unexpected error: %v\n", err)
					} else if string(resMapYaml) != expectedK8sYaml {
						t.Fatalf("expected k8s yaml: [%v] but got: [%v]\n", expectedK8sYaml, string(resMapYaml))
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
	}

	testExecutableResolver := &testExecutableResolverT{}
	if kustomizeExecutablePath, err := exec.LookPath("kustomize"); err != nil {
		tmpDirKustomizeExecutablePath := filepath.Join(os.TempDir(), "kustomize")
		if info, err := os.Stat(tmpDirKustomizeExecutablePath); err == nil && info.Mode().IsRegular() {
			testExecutableResolver.path = tmpDirKustomizeExecutablePath
		} else if _, err := downloadLatestKustomizeExecutable(os.TempDir()); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		} else {
			testExecutableResolver.path = tmpDirKustomizeExecutablePath
		}
	} else {
		testExecutableResolver.path = kustomizeExecutablePath
	}
	plugin := GoGetterPlugin{logger: log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds), executableResolver: testExecutableResolver}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)

			tmpPluginHomeDir := filepath.Join(testCase.loaderRootDir, "plugin_home")
			if err := os.Mkdir(tmpPluginHomeDir, os.ModePerm); err != nil {
				t.Fatalf("Err: %v", err)
			} else if err := os.Setenv("KUSTOMIZE_PLUGIN_HOME", tmpPluginHomeDir); err != nil {
				t.Fatalf("Err: %v", err)
			}

			ldr, err := loader.NewLoader(loader.RestrictionRootOnly, testCase.loaderRootDir, filesys.MakeFsOnDisk())
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			h := resmap.NewPluginHelpers(ldr, valtest_test.MakeHappyMapValidator(t), resourceFactory)
			if err := plugin.Config(h, []byte(testCase.pluginConfig)); err != nil {
				t.Fatalf("Err: %v", err)
			}

			if resMap, err := plugin.Generate(); err != nil {
				t.Fatalf("Err: %v", err)
			} else {
				testCase.checkAssertions(t, resMap)
			}
		})
	}
}

func downloadLatestKustomizeExecutable(destDir string) (string, error) {
	apiResp, err := http.Get("https://api.github.com/repos/qlik-oss/kustomize/releases/latest")
	if err != nil {
		return "", err
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid api response code: %v", apiResp.StatusCode)
	}

	buff := &bytes.Buffer{}
	if _, err = io.Copy(buff, apiResp.Body); err != nil {
		return "", err
	}

	mapReleaseInfo := make(map[string]interface{})
	if err := json.Unmarshal(buff.Bytes(), &mapReleaseInfo); err != nil {
		return "", err
	}

	archiveDownloadUrl := ""
	if assets, ok := mapReleaseInfo["assets"].([]interface{}); !ok {
		return "", errors.New("unable to extract the release assets slice")
	} else {
		for _, asset := range assets {
			if assetMap, ok := asset.(map[string]interface{}); !ok {
				return "", errors.New("unable to extract the release asset")
			} else if url, ok := assetMap["browser_download_url"].(string); !ok {
				return "", errors.New("unable to extract the release asset's browser_download_url")
			} else if strings.Contains(url, runtime.GOOS) {
				archiveDownloadUrl = url
				break
			}
		}
	}

	if archiveDownloadUrl == "" {
		return "", fmt.Errorf("unable to extract download URL for the current runtime: %v", runtime.GOOS)
	}

	downloadResp, err := http.Get(archiveDownloadUrl)
	if err != nil {
		return "", err
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid download response code: %v", downloadResp.StatusCode)
	}

	archiveName := filepath.Base(archiveDownloadUrl)
	archivePath := filepath.Join(destDir, archiveName)
	f, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err = io.Copy(f, downloadResp.Body); err != nil {
		return "", err
	} else if err := os.Chmod(archivePath, os.ModePerm); err != nil {
		return "", err
	} else if err := archiver.Unarchive(archivePath, destDir); err != nil {
		return "", err
	} else if err := os.Chmod(filepath.Join(destDir, "kustomize"), os.ModePerm); err != nil {
		return "", err
	}

	return filepath.Join(destDir, "kustomize"), nil
}
