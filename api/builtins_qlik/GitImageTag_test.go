package builtins_qlik

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

func Test_ImageGitTag_getHighestSemverTag(t *testing.T) {
	type tcT struct {
		name                  string
		testVersions          []string
		expectedLatestVersion string
	}
	testCases := []*tcT{
		func() *tcT {
			testVersions := make([]string, 0)
			for i := 1; i <= 10; i++ {
				testVersions = append(testVersions, fmt.Sprintf("v0.0.%v", i))
			}
			return &tcT{
				name:                  "all good semvers",
				testVersions:          testVersions,
				expectedLatestVersion: testVersions[len(testVersions)-1],
			}
		}(),
		func() *tcT {
			testVersions := make([]string, 0)
			for i := 1; i <= 10; i++ {
				testVersions = append(testVersions, fmt.Sprintf("v0.0.%v", i))
			}
			return &tcT{
				name:                  "all bad semvers",
				testVersions:          []string{"foo", "bar", "baz"},
				expectedLatestVersion: "",
			}
		}(),
		func() *tcT {
			testVersions := make([]string, 0)
			for i := 1; i <= 10; i++ {
				testVersions = append(testVersions, fmt.Sprintf("v0.0.%v", i))
			}
			expected := "v1.2.3-rc1-with-hypen"
			return &tcT{
				name:                  "some good and some bad semvers",
				testVersions:          []string{"v1.0.0", "foo", expected},
				expectedLatestVersion: expected,
			}
		}(),
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			latestSemverTag, err := getHighestSemverTag(testCase.testVersions, log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds))
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			if latestSemverTag != testCase.expectedLatestVersion {
				t.Fatalf("expected: %v, but got: %v\n", testCase.expectedLatestVersion, latestSemverTag)
			}
		})
	}
}

func Test_ImageGitTag_getGitTagsForHead(t *testing.T) {
	type tcT struct {
		name     string
		dir      string
		validate func(t *testing.T, tags []string)
	}

	testCases := []*tcT{
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			if _, err := execCmd(tmpDir, "git", "init"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			} else if _, err := execCmd(tmpDir, "git", "config", "user.email", "you@example.com"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			numTags := 10
			for i := 0; i <= numTags; i++ {
				if err := writeAndCommitFile(tmpDir, fmt.Sprintf("foo-%v.txt", i), fmt.Sprintf("foo-%v", i), fmt.Sprintf("committing foo-%v.txt", i)); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				} else if _, err := execCmd(tmpDir, "git", "tag", fmt.Sprintf("v0.0.%v", i)); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}
			}

			return &tcT{
				name: "single tag",
				dir:  tmpDir,
				validate: func(t *testing.T, tags []string) {
					expectedTags := []string{fmt.Sprintf("v0.0.%v", numTags)}
					if !reflect.DeepEqual(tags, expectedTags) {
						t.Fatalf("expected: %v, but got: %v\n", expectedTags, tags)
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

			if _, err := execCmd(tmpDir, "git", "init"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			} else if _, err := execCmd(tmpDir, "git", "config", "user.email", "you@example.com"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			numTags := 10
			for i := 0; i <= numTags; i++ {
				if err := writeAndCommitFile(tmpDir, fmt.Sprintf("foo-%v.txt", i), fmt.Sprintf("foo-%v", i), fmt.Sprintf("committing foo-%v.txt", i)); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				} else if _, err := execCmd(tmpDir, "git", "tag", fmt.Sprintf("v0.0.%v-foo", i)); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				} else if _, err := execCmd(tmpDir, "git", "tag", fmt.Sprintf("v0.0.%v-bar", i)); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}
			}

			return &tcT{
				name: "multiple tags",
				dir:  tmpDir,
				validate: func(t *testing.T, tags []string) {
					foundTagsMap := make(map[string]bool)
					for _, tag := range tags {
						foundTagsMap[tag] = true
					}
					expectedTags := map[string]bool{
						fmt.Sprintf("v0.0.%v-foo", numTags): true,
						fmt.Sprintf("v0.0.%v-bar", numTags): true,
					}

					if !reflect.DeepEqual(foundTagsMap, expectedTags) {
						t.Fatalf("expected: %v, but got: %v\n", expectedTags, tags)
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

			if _, err := execCmd(tmpDir, "git", "init"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			} else if _, err := execCmd(tmpDir, "git", "config", "user.email", "you@example.com"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			barDir := filepath.Join(tmpDir, "bar-dir")
			if err := writeAndCommitFile(tmpDir, "foo.txt", "foo", "committing foo.txt"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			} else if err := os.MkdirAll(barDir, os.ModePerm); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			} else if err := writeAndCommitFile(barDir, "bar.txt", "bar", "committing bar.txt"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			} else if _, err := execCmd(tmpDir, "git", "tag", "foobar"); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "git operations from subdirectory",
				dir:  barDir,
				validate: func(t *testing.T, tags []string) {
					expectedTags := []string{"foobar"}
					if !reflect.DeepEqual(tags, expectedTags) {
						t.Fatalf("expected: %v, but got: %v\n", expectedTags, tags)
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tags, err := getGitTagsForHead(testCase.dir)
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			testCase.validate(t, tags)
		})
	}
}

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

			subDir, err := setupGitDirWithSubdir(tmpDir, []string{"foobar"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "non-semver tag",
				dir:  subDir,
				validate: func(t *testing.T, tag string) {
					if tag != "" {
						t.Fatalf("expected: %v, but got: %v\n", "", tag)
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
			subDir, err := setupGitDirWithSubdir(tmpDir, []string{"foobar", semverTag})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "one semver tag ",
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
			subDir, err := setupGitDirWithSubdir(tmpDir, []string{"foo", "v1.0.0", "bar", highestSemverTag, "baz", "v0.0.1", "boo", "v100.0.0-alpha"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "highest semver tag ",
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

func setupGitDirWithSubdir(tmpDir string, tags []string) (string, error) {
	if _, err := execCmd(tmpDir, "git", "init"); err != nil {
		return "", err
	} else if _, err := execCmd(tmpDir, "git", "config", "user.email", "you@example.com"); err != nil {
		return "", err
	}

	barDir := filepath.Join(tmpDir, "bar-dir")
	if err := writeAndCommitFile(tmpDir, "foo.txt", "foo", "committing foo.txt"); err != nil {
		return "", err
	} else if err := os.MkdirAll(barDir, os.ModePerm); err != nil {
		return "", err
	} else if err := writeAndCommitFile(barDir, "bar.txt", "bar", "committing bar.txt"); err != nil {
		return "", err
	} else {
		for _, tag := range tags {
			if _, err := execCmd(tmpDir, "git", "tag", tag); err != nil {
				return "", err
			}
		}
		return barDir, nil
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

			semverTag := "v0.0.1"
			subDir, err := setupGitDirWithSubdir(tmpDir, []string{"foo", semverTag, "bar"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "single semver git tag chosen",
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

			highestSemverTag := "v5.0.0"
			subDir, err := setupGitDirWithSubdir(tmpDir, []string{"foo", "1.0.0", highestSemverTag, "bar", "2.0.0"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "highest semver git tag chosen",
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
					expected := fmt.Sprintf(outputResourcesTemplate, highestSemverTag, highestSemverTag, highestSemverTag, highestSemverTag)

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

			highestSemverTag := "v5.0.0"
			subDir, err := setupGitDirWithSubdir(tmpDir, []string{"foo", "1.0.0", highestSemverTag, "bar", "2.0.0"})
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
		func() *tcT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			defaultSemverTag := "v5.0.0"
			subDir, err := setupGitDirWithSubdir(tmpDir, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "no semver tag, default used",
				pluginConfig: fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: GitImageTag
metadata:
  name: notImportantHere
images:
  - name: nginx
    default: %v
`, defaultSemverTag),
				pluginInputResources: pluginInputResources,
				loaderRootDir:        subDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expected := fmt.Sprintf(outputResourcesTemplate, defaultSemverTag, defaultSemverTag, defaultSemverTag, defaultSemverTag)

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

			subDir, err := setupGitDirWithSubdir(tmpDir, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return &tcT{
				name: "no semver tag and no default = no change",
				pluginConfig: `
apiVersion: qlik.com/v1
kind: GitImageTag
metadata:
 name: notImportantHere
name: nginx
`,
				pluginInputResources: pluginInputResources,
				loaderRootDir:        subDir,
				checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
					expected := pluginInputResources

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
