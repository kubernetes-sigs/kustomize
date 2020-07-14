package builtins_qlik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func Test_ImageGitTag_getHighestSemverGitTagForHead(t *testing.T) {
	type tcT struct {
		name     string
		dir      string
		validate func(t *testing.T, tag string)
	}

	testCases := []*tcT{
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			subDir, shortGitRef, err := setupGitDirWithSubdir(tmpDir, []string{"foobar"}, []string{"foo-tag"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "no semver tags",
				dir:  subDir,
				validate: func(t *testing.T, tag string) {
					expected := fmt.Sprintf("v0.0.0-%v", shortGitRef)
					if tag != expected {
						t.Fatalf("expected: %v, but got: %v\n", expected, tag)
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			semverTag := "v1.0.0"
			subDir, shortGitRef, err := setupGitDirWithSubdir(tmpDir, []string{}, []string{semverTag})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "semver tag before head",
				dir:  subDir,
				validate: func(t *testing.T, tag string) {
					expected := fmt.Sprintf("%v-%v", semverTag, shortGitRef)
					if tag != expected {
						t.Fatalf("expected: %v, but got: %v\n", expected, tag)
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			semverTag := "v1.0.0"
			subDir, _, err := setupGitDirWithSubdir(tmpDir, []string{"foobar", semverTag}, []string{"foo-tag"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "semver tag on head",
				dir:  subDir,
				validate: func(t *testing.T, tag string) {
					if tag != semverTag {
						t.Fatalf("expected: %v, but got: %v\n", semverTag, tag)
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			highestSemverTag := "v100.0.0-beta"
			subDir, _, err := setupGitDirWithSubdir(tmpDir, []string{"foo", "v1.0.0", "bar", highestSemverTag, "baz", "v0.0.1", "boo", "v100.0.0-alpha"}, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "highest semver tag",
				dir:  subDir,
				validate: func(t *testing.T, tag string) {
					if tag != highestSemverTag {
						t.Fatalf("expected: %v, but got: %v\n", highestSemverTag, tag)
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tag, err := getHighestSemverGitTagForHead(testCase.dir, log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds))
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			testCase.validate(t, tag)
		})
	}
}

func execCmd(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Output()
}

func writeAndCommitFile(dir, fileName, fileContent, commitMessage string) error {
	if err := ioutil.WriteFile(filepath.Join(dir, fileName), []byte(fileContent), os.ModePerm); err != nil {
		return errors.Wrapf(err, "error writing file: %v", filepath.Join(dir, fileName))
	} else if _, err := execCmd(dir, "git", "add", "."); err != nil {
		return errors.Wrap(err, "error executing git add")
	} else if _, err := execCmd(dir, "git", "commit", "-m", commitMessage); err != nil {
		return errors.Wrap(err, "error executing git commit")
	}
	return nil
}

func setupGitDirWithSubdir(tmpDir string, headTags []string, intermediateTags []string) (dir string, shortGitRef string, err error) {
	if _, err := execCmd(tmpDir, "git", "init"); err != nil {
		return "", "", err
	} else if _, err := execCmd(tmpDir, "git", "config", "user.email", "you@example.com"); err != nil {
		return "", "", err
	}

	barDir := filepath.Join(tmpDir, "bar-dir")
	if err := writeAndCommitFile(tmpDir, "foo.txt", "foo", "committing foo.txt"); err != nil {
		return "", "", err
	} else {
		for _, tag := range intermediateTags {
			if _, err := execCmd(tmpDir, "git", "tag", tag); err != nil {
				return "", "", err
			}
		}
	}

	if err := os.MkdirAll(barDir, os.ModePerm); err != nil {
		return "", "", err
	} else if err := writeAndCommitFile(barDir, "bar.txt", "bar", "committing bar.txt"); err != nil {
		return "", "", err
	} else {
		for _, tag := range headTags {
			if _, err := execCmd(tmpDir, "git", "tag", tag); err != nil {
				return "", "", err
			}
		}
	}

	if shortGitRefBytes, err := execCmd(tmpDir, "git", "rev-parse", "--short", "HEAD"); err != nil {
		return "", "", err
	} else {
		return barDir, string(bytes.TrimSpace(shortGitRefBytes)), nil
	}
}

func Test_ImageGitTag_Transform(t *testing.T) {
	type tcT struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		loaderRootDir        string
		checkAssertions      func(*testing.T, resmap.ResMap)
	}

	pluginInputResources := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`
	outputResourcesTemplate := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:%v
        name: nginx-tagged
      - image: nginx:%v
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx:%v
        name: nginx-notag
      - image: nginx:%v
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`
	testCases := []*tcT{
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			subDir, hash, err := setupGitDirWithSubdir(tmpDir, []string{}, []string{"foo"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "no semver tags",
				pluginConfig: `
apiVersion: qlik.com/v1
kind: GitImageTag
metadata:
  name: notImportantHere
images:
  - name: nginx
`,
				pluginInputResources: pluginInputResources,
				loaderRootDir:        subDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expectedTag := fmt.Sprintf("v0.0.0-%v", hash)
					expected := fmt.Sprintf(outputResourcesTemplate, expectedTag, expectedTag, expectedTag, expectedTag)

					actual, err := resMap.AsYaml()
					if err != nil {
						t.Fatalf("Err: %v", err)
					} else if string(actual) != expected {
						t.Fatalf("expected:\n%v\n, but got:\n%v", expected, string(actual))
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			semverTag := "v0.0.1"
			subDir, hash, err := setupGitDirWithSubdir(tmpDir, []string{}, []string{"foo", semverTag, "bar"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "semver tag before head",
				pluginConfig: `
apiVersion: qlik.com/v1
kind: GitImageTag
metadata:
  name: notImportantHere
images:
  - name: nginx
`,
				pluginInputResources: pluginInputResources,
				loaderRootDir:        subDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expectedTag := fmt.Sprintf("%v-%v", semverTag, hash)
					expected := fmt.Sprintf(outputResourcesTemplate, expectedTag, expectedTag, expectedTag, expectedTag)

					actual, err := resMap.AsYaml()
					if err != nil {
						t.Fatalf("Err: %v", err)
					} else if string(actual) != expected {
						t.Fatalf("expected:\n%v\n, but got:\n%v", expected, string(actual))
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			semverTag := "v1.0.2"
			subDir, _, err := setupGitDirWithSubdir(tmpDir, []string{semverTag, "v0.0.2"}, []string{"foo", "v0.0.1", "bar"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "semver tag on head",
				pluginConfig: `
apiVersion: qlik.com/v1
kind: GitImageTag
metadata:
  name: notImportantHere
images:
  - name: nginx
`,
				pluginInputResources: pluginInputResources,
				loaderRootDir:        subDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expected := fmt.Sprintf(outputResourcesTemplate, semverTag, semverTag, semverTag, semverTag)

					actual, err := resMap.AsYaml()
					if err != nil {
						t.Fatalf("Err: %v", err)
					} else if string(actual) != expected {
						t.Fatalf("expected:\n%v\n, but got:\n%v", expected, string(actual))
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			highestSemverTag := "5.0.0"
			subDir, _, err := setupGitDirWithSubdir(tmpDir, []string{"1.0.0", highestSemverTag, "bar", "2.0.0"}, []string{"foo", "v0.0.1"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "highest semver git tag chosen and used for multiple images",
				pluginConfig: `
apiVersion: qlik.com/v1
kind: GitImageTag
metadata:
  name: notImportantHere
images:
  - name: nginx
  - name: postgres
  - name: alpine
`,
				pluginInputResources: pluginInputResources,
				loaderRootDir:        subDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expected := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:%v
        name: nginx-tagged
      - image: nginx:%v
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:%v
        name: postgresdb
      initContainers:
      - image: nginx:%v
        name: nginx-notag
      - image: nginx:%v
        name: nginx-sha256
      - image: alpine:%v
        name: init-alpine
`, highestSemverTag, highestSemverTag, highestSemverTag,
						highestSemverTag, highestSemverTag, highestSemverTag)

					actual, err := resMap.AsYaml()
					if err != nil {
						t.Fatalf("Err: %v", err)
					} else if string(actual) != expected {
						t.Fatalf("expected:\n%v\n, but got:\n%v", expected, string(actual))
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
	}
	plugin := GitImageTagPlugin{logger: log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds)}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
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

			if err := plugin.Transform(resMap); err != nil {
				t.Fatalf("Err: %v", err)
			}

			testCase.checkAssertions(t, resMap)
		})
	}
}
