package builtins_qlik

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mholt/archiver/v3"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
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
		checkAssertions func(*testing.T, resmap.ResMap)
	}
	testCases := []*tcT{
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			kustomizationYamlFilePath := path.Join(tmpDir, "kustomization.yaml")
			kustomizationYaml := `
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
- name: foo-config
  literals:    
  - foo=bar
`
			err = ioutil.WriteFile(kustomizationYamlFilePath, []byte(kustomizationYaml), os.ModePerm)
			if err != nil {
				t.Fatalf("error writing kustomization file to path: %v error: %v\n", kustomizationYamlFilePath, err)
			}

			return &tcT{
				name: "go-get and kustomize",
				pluginConfig: fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: GoGetter
metadata:
  name: notImportantHere
url: %s
`, kustomizationYamlFilePath),
				loaderRootDir: tmpDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expectedK8sYaml := `apiVersion: v1
data:
  foo: bar
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
			resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

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
