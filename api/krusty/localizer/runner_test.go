// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/api/krusty/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	customSchema = `{
  "definitions": {
    "v1alpha1.MyCRD": {
      "properties": {
        "apiVersion": {
          "type": "string"
        },
        "kind": {
          "type": "string"
        },
        "metadata": {
          "type": "object"
        },
        "spec": {
          "properties": {
            "template": {
              "$ref": "#/definitions/io.k8s.api.core.v1.PodTemplateSpec"
            }
          },
          "type": "object"
        },
        "status": {
           "properties": {
            "success": {
              "type": "boolean"
            }
          },
          "type": "object"
        }
      },
      "type": "object",
      "x-kubernetes-group-version-kind": [
        {
          "group": "example.com",
          "kind": "MyCRD",
          "version": "v1alpha1"
        },
        {
          "group": "",
          "kind": "MyCRD",
          "version": "v1alpha1"
        }
      ]
    },
    "io.k8s.api.core.v1.PodTemplateSpec": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/io.k8s.api.core.v1.PodSpec"
        }
      },
      "type": "object"
    },
    "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta": {
      "properties": {
        "name": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "io.k8s.api.core.v1.PodSpec": {
      "properties": {
        "containers": {
          "items": {
            "$ref": "#/definitions/io.k8s.api.core.v1.Container"
          },
          "type": "array",
          "x-kubernetes-patch-merge-key": "name",
          "x-kubernetes-patch-strategy": "merge"
        }
      },
      "type": "object"
    },
    "io.k8s.api.core.v1.Container": {
      "properties": {
        "command": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "image": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "ports": {
         "items": {
            "$ref": "#/definitions/io.k8s.api.core.v1.ContainerPort"
          },
          "type": "array",
          "x-kubernetes-list-map-keys": [
            "containerPort",
            "protocol"
          ],
          "x-kubernetes-list-type": "map",
          "x-kubernetes-patch-merge-key": "containerPort",
          "x-kubernetes-patch-strategy": "merge"
        }
      },
      "type": "object"
    },
    "io.k8s.api.core.v1.ContainerPort": {
     "properties": {
        "containerPort": {
          "format": "int32",
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "protocol": {
          "type": "string"
        }
      },
      "type": "object"
    }
  }
}
`

	simpleURL = "https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/simple"

	simpleKustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: localize-
resources:
- deployment.yaml
- service.yaml
`

	simpleDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment-simple
  labels:
    app: deployment-simple
spec:
  selector:
    matchLabels:
      app: simple
  template:
    metadata:
      labels:
        app: simple
    spec:
      containers:
      - name: nginx
        image: nginx:1.16
        ports:
        - containerPort: 8080
`
	simpleService = `apiVersion: v1
kind: Service
metadata:
  name: test-service-simple
spec:
  selector:
    app: deployment-simple
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080 
`

	remoteHPA = `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: hpa-deployment
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: localize-test-deployment-simple
  minReplicas: 1
  maxReplicas: 10`

	urlQuery = "?submodules=0&ref=kustomize/v4.5.7&timeout=300"
)

func prepareFs(t *testing.T, dirs []string, files map[string]string) (
	memoryFs filesys.FileSystem, actualFs filesys.FileSystem, testDir filesys.ConfirmedDir) {
	t.Helper()

	memoryFs = filesys.MakeFsInMemory()
	actualFs = filesys.MakeFsOnDisk()

	testDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)

	setupDir(t, memoryFs, testDir.String(), files)
	for _, dirPath := range dirs {
		require.NoError(t, actualFs.MkdirAll(testDir.Join(dirPath)))
	}
	setupDir(t, actualFs, testDir.String(), files)

	t.Cleanup(func() {
		_ = actualFs.RemoveAll(testDir.String())
	})

	return memoryFs, actualFs, testDir
}

func setupDir(t *testing.T, targetFs filesys.FileSystem, parentDir string, files map[string]string) {
	t.Helper()

	for file, content := range files {
		require.NoError(t, targetFs.WriteFile(filepath.Join(parentDir, file), []byte(content)))
	}
}

func setWorkingDir(t *testing.T, workingDir string) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	err = os.Chdir(workingDir)
	require.NoError(t, err)
}

func link(t *testing.T, testDir filesys.ConfirmedDir, links map[string]string) {
	t.Helper()

	for newLink, file := range links {
		require.NoError(t, os.Symlink(testDir.Join(file), testDir.Join(newLink)))
	}
}

func simplePathAndFiles(t *testing.T) (locPath string, files map[string]string) {
	t.Helper()

	locPath = filepath.Join(LocalizeDir, "github.com",
		"kubernetes-sigs", "kustomize", "kustomize", "v4.5.7",
		"api", "krusty", "testdata", "localize", "simple")
	files = map[string]string{
		"kustomization.yaml": simpleKustomization,
		"deployment.yaml":    simpleDeployment,
		"service.yaml":       simpleService,
	}
	return
}

func remotePathAndFiles(t *testing.T) (locPath string, files map[string]string) {
	t.Helper()

	locPath = filepath.Join(LocalizeDir, "github.com",
		"kubernetes-sigs", "kustomize", "master",
		"api", "krusty", "testdata", "localize", "remote")
	simplePath, simpleFiles := simplePathAndFiles(t)
	files = map[string]string{
		"kustomization.yaml": fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  purpose: remoteReference
kind: Kustomization
resources:
- %s
- hpa.yaml
`, simplePath),
		"hpa.yaml": remoteHPA,
	}
	for path, content := range simpleFiles {
		files[filepath.Join(simplePath, path)] = content
	}
	return
}

// checkFs checks fsActual, the real file system, against fsExpected, a file system in memory, for contents
// in directory walkDir. checkFs does not allow symlinks.
func checkFs(t *testing.T, walkDir string, fsExpected filesys.FileSystem, fsActual filesys.FileSystem) {
	t.Helper()

	err := fsActual.Walk(walkDir, func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		require.NotEqual(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)
		require.True(t, fsExpected.Exists(path), "unexpected file %q", path)
		return nil
	})
	require.NoError(t, err)

	err = fsExpected.Walk(walkDir, func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		if info.IsDir() {
			require.DirExists(t, path)
		} else {
			require.FileExists(t, path)

			expectedContent, err := fsExpected.ReadFile(path)
			require.NoError(t, err)
			actualContent, err := fsActual.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, string(expectedContent), string(actualContent))
		}
		return nil
	})
	require.NoError(t, err)
}

func TestWorkingDir(t *testing.T) {
	files := map[string]string{
		filepath.Join("target", "kustomization.yaml"): fmt.Sprintf(`resources:
- %s
`, filepath.Join("..", "base")),
		filepath.Join("base", "kustomization.yaml"): `resources:
- deployment.yaml
`,
		filepath.Join("base", "deployment.yaml"): simpleDeployment,
	}
	fsExpected, fsActual, wd := prepareFs(t, []string{"target", "base"}, files)
	setWorkingDir(t, wd.String())

	err := localizer.Run(fsActual, "target", ".", "")
	require.NoError(t, err)

	dst := wd.Join("localized-target")
	setupDir(t, fsExpected, dst, files)
	checkFs(t, dst, fsExpected, fsActual)
}

func TestSymlinks(t *testing.T) {
	// test directory
	// - link to target
	// - link to base
	// - link to file
	// - target (and scope)
	//   - link to kustomization
	//   - base
	//   - nested root
	//     - file
	//     - kustomization
	fsExpected, fsActual, testDir := prepareFs(t, []string{"target",
		filepath.Join("target", "base"),
		filepath.Join("target", "nested")}, map[string]string{
		filepath.Join("target", "base", "kustomization.yaml"): `namePrefix: test-
`,
		filepath.Join("target", "nested", "kustomization"): fmt.Sprintf(`resources:
- %s
- %s
`, filepath.Join("..", "file-link"), filepath.Join("..", "base-link")),
		filepath.Join("target", "nested", "file"): simpleDeployment,
	})
	link(t, testDir, map[string]string{
		"target-link": "target",
		"base-link":   filepath.Join("target", "base"),
		"file-link":   filepath.Join("target", "nested", "file"),
		filepath.Join("target", "kustomization.yaml"): filepath.Join("target", "nested", "kustomization"),
	})
	setWorkingDir(t, testDir.String())

	err := localizer.Run(fsActual, "target-link", "target", "")
	require.NoError(t, err)

	dst := testDir.Join("localized-target")
	setupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
- base
`, filepath.Join("nested", "file")),
		filepath.Join("base", "kustomization.yaml"): `namePrefix: test-
`,
		filepath.Join("nested", "file"): simpleDeployment,
	})
	checkFs(t, dst, fsExpected, fsActual)
}

func TestRemoteTargetDefaultDst(t *testing.T) {
	fsExpected, fsActual, testDir := prepareFs(t, nil, nil)
	setWorkingDir(t, testDir.String())

	const target = simpleURL + urlQuery
	err := localizer.Run(fsActual, target, "", "")
	require.NoError(t, err)

	dst := testDir.Join("localized-simple-kustomize-v4.5.7")
	_, files := simplePathAndFiles(t)
	setupDir(t, fsExpected,
		filepath.Join(dst, "api", "krusty", "testdata", "localize", "simple"),
		files)
	checkFs(t, testDir.String(), fsExpected, fsActual)
}

func TestBadArgs(t *testing.T) {
	badDst := filepath.Join("non-existing", "dst")

	for name, test := range map[string]struct {
		target string
		scope  string
		dst    string
		err    string
	}{
		"target_no_ref": {
			target: simpleURL,
			err:    `localize remote root "https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/simple" missing ref query string parameter`,
		},
		"non-empty_scope": {
			target: simpleURL + urlQuery,
			scope:  ".",
			err:    fmt.Sprintf(`invalid localize scope ".": scope "." specified for remote localize target "%s"`, simpleURL+urlQuery),
		},
		"dst_in_non-existing_dir": {
			target: ".",
			dst:    badDst,
			err:    fmt.Sprintf(`invalid localize destination "%s": unable to create localize destination directory: mkdir %s: no such file or directory`, badDst, badDst),
		},
	} {
		t.Run(name, func(t *testing.T) {
			kust := map[string]string{
				"kustomization.yaml": "namePrefix: test-",
			}
			fsExpected, fsActual, testDir := prepareFs(t, nil, kust)
			setWorkingDir(t, testDir.String())

			err := localizer.Run(fsActual, test.target, test.scope, test.dst)
			require.EqualError(t, err, test.err)

			setupDir(t, fsExpected, testDir.String(), kust)
			checkFs(t, testDir.String(), fsExpected, fsActual)
		})
	}
}

func TestRemoteFile(t *testing.T) {
	const kustf = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
openapi:
  path: %s
`
	fsExpected, fsActual, testDir := prepareFs(t, nil, map[string]string{
		"kustomization.yaml": fmt.Sprintf(kustf, `https://raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/customschema.json`),
	})

	dst := testDir.Join("dst")
	err := localizer.Run(fsActual, testDir.String(), "", dst)
	require.NoError(t, err)

	localizedPath := filepath.Join(LocalizeDir, "raw.githubusercontent.com",
		"kubernetes-sigs", "kustomize", "kustomize", "v4.5.7", "api", "krusty",
		"testdata", "customschema.json")
	setupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(kustf, localizedPath),
		localizedPath:        customSchema,
	})
	checkFs(t, testDir.String(), fsExpected, fsActual)
}

func TestRemoteRoot(t *testing.T) {
	fsExpected, fsActual, testDir := prepareFs(t, nil, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, simpleURL+urlQuery),
	})

	dst := testDir.Join("dst")
	err := localizer.Run(fsActual, testDir.String(), "", dst)
	require.NoError(t, err)

	localizedPath, files := simplePathAndFiles(t)
	setupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, localizedPath),
	})
	setupDir(t, fsExpected, filepath.Join(dst, localizedPath), files)
	checkFs(t, dst, fsExpected, fsActual)
}

func TestNestedRemoteRoots(t *testing.T) {
	fsExpected, fsActual, testDir := prepareFs(t, nil, map[string]string{
		// TODO(annasong): Change the ref to the release after kustomize/v4.5.7.
		// We need changes to remote post-kustomize/v4.5.7.
		"kustomization.yaml": `resources:
- https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/remote?submodules=0&ref=master&timeout=300
`,
	})

	dst := testDir.Join("dst")
	err := localizer.Run(fsActual, testDir.String(), "", dst)
	require.NoError(t, err)

	localizedPath, files := remotePathAndFiles(t)
	setupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, localizedPath),
	})
	setupDir(t, fsExpected, filepath.Join(dst, localizedPath), files)
	checkFs(t, dst, fsExpected, fsActual)
}

func TestResourcesRepoNotFile(t *testing.T) {
	const repo = "https://github.com/kubernetes-sigs/kustomize" + urlQuery
	kustomization := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, repo),
	}
	fsExpected, fsActual, testDir := prepareFs(t, nil, kustomization)

	err := localizer.Run(fsActual, testDir.String(), "", testDir.Join("dst"))

	const readmeErr = `yaml: line 28: mapping values are not allowed in this context`
	fileErr := fmt.Sprintf(`invalid resource at file "%s": MalformedYAMLError: %s`, repo, readmeErr)
	rootErr := fmt.Sprintf(`unable to localize root "%s": unable to find one of 'kustomization.yaml', 'kustomization.yml' or 'Kustomization'`, repo)
	var actualErr PathLocalizeError
	require.ErrorAs(t, err, &actualErr)
	require.Equal(t, repo, actualErr.Path)
	require.EqualError(t, actualErr.FileError, fileErr)
	require.ErrorContains(t, actualErr.RootError, rootErr)

	setupDir(t, fsExpected, testDir.String(), kustomization)
	checkFs(t, testDir.String(), fsExpected, fsActual)
}

func TestRemoteRootNoRef(t *testing.T) {
	const root = simpleURL + "?submodules=0&timeout=300"
	kustomization := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, root),
	}
	fsExpected, fsActual, testDir := prepareFs(t, nil, kustomization)

	err := localizer.Run(fsActual, testDir.String(), "", testDir.Join("dst"))

	const fileErr = "invalid file reference: URL is a git repository"
	rootErr := fmt.Sprintf(`localize remote root "%s" missing ref query string parameter`, root)
	var actualErr PathLocalizeError
	require.ErrorAs(t, err, &actualErr)
	require.Equal(t, root, actualErr.Path)
	require.EqualError(t, actualErr.FileError, fileErr)
	require.EqualError(t, actualErr.RootError, rootErr)

	setupDir(t, fsExpected, testDir.String(), kustomization)
	checkFs(t, testDir.String(), fsExpected, fsActual)
}

func TestExistingCacheDir(t *testing.T) {
	const remoteFile = `https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/deployment.yaml`
	file := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, remoteFile),
		filepath.Join(LocalizeDir, "file"): "existing",
	}
	fsExpected, fsActual, testDir := prepareFs(t, []string{LocalizeDir}, file)

	err := localizer.Run(fsActual, testDir.String(), "", testDir.Join("dst"))
	require.ErrorContains(t, err, fmt.Sprintf(`already contains localized-files needed to store file "%s"`, remoteFile))

	setupDir(t, fsExpected, testDir.String(), file)
	checkFs(t, testDir.String(), fsExpected, fsActual)
}
