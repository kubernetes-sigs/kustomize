// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/api/krusty/localizer"
	. "sigs.k8s.io/kustomize/api/testutils/localizertest"
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

	valuesFile = `minecraftServer:
  difficulty: peaceful
`
)

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
	fsExpected, fsActual, wd := PrepareFs(t, []string{"target", "base"}, files)
	SetWorkingDir(t, wd.String())

	dst, err := localizer.Run(fsActual, "target", ".", "")
	require.NoError(t, err)
	require.Equal(t, wd.Join("localized-target"), dst)

	SetupDir(t, fsExpected, dst, files)
	CheckFs(t, wd.String(), fsExpected, fsActual)
}

func TestLoaderSymlinks(t *testing.T) {
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
	fsExpected, fsActual, testDir := PrepareFs(t, []string{"target",
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
	SetWorkingDir(t, testDir.String())

	dst, err := localizer.Run(fsActual, "target-link", "target", "")
	require.NoError(t, err)
	require.Equal(t, testDir.Join("localized-target"), dst)

	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
- base
`, filepath.Join("nested", "file")),
		filepath.Join("base", "kustomization.yaml"): `namePrefix: test-
`,
		filepath.Join("nested", "file"): simpleDeployment,
	})
	CheckFs(t, dst, fsExpected, fsActual)
}

func TestRemoteTargetDefaultDst(t *testing.T) {
	fsExpected, fsActual, testDir := PrepareFs(t, nil, nil)
	SetWorkingDir(t, testDir.String())

	const target = simpleURL + urlQuery
	dst, err := localizer.Run(fsActual, target, "", "")
	require.NoError(t, err)
	require.Equal(t, testDir.Join("localized-simple-kustomize-v4.5.7"), dst)

	_, files := simplePathAndFiles(t)
	SetupDir(t, fsExpected,
		filepath.Join(dst, "api", "krusty", "testdata", "localize", "simple"),
		files)
	CheckFs(t, testDir.String(), fsExpected, fsActual)
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
			fsExpected, fsActual, testDir := PrepareFs(t, nil, kust)
			SetWorkingDir(t, testDir.String())

			_, err := localizer.Run(fsActual, test.target, test.scope, test.dst)
			require.EqualError(t, err, test.err)

			SetupDir(t, fsExpected, testDir.String(), kust)
			CheckFs(t, testDir.String(), fsExpected, fsActual)
		})
	}
}

func TestRemoteFile(t *testing.T) {
	const kustf = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
openapi:
  path: %s
`
	fsExpected, fsActual, testDir := PrepareFs(t, nil, map[string]string{
		"kustomization.yaml": fmt.Sprintf(kustf, `https://raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/customschema.json`),
	})

	newDir := testDir.Join("dst")
	dst, err := localizer.Run(fsActual, testDir.String(), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	localizedPath := filepath.Join(LocalizeDir, "raw.githubusercontent.com",
		"kubernetes-sigs", "kustomize", "kustomize", "v4.5.7", "api", "krusty",
		"testdata", "customschema.json")
	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(kustf, localizedPath),
		localizedPath:        customSchema,
	})
	CheckFs(t, testDir.String(), fsExpected, fsActual)
}

func TestRemoteRoot(t *testing.T) {
	fsExpected, fsActual, testDir := PrepareFs(t, nil, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, simpleURL+urlQuery),
	})

	newDir := testDir.Join("dst")
	dst, err := localizer.Run(fsActual, testDir.String(), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	localizedPath, files := simplePathAndFiles(t)
	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, localizedPath),
	})
	SetupDir(t, fsExpected, filepath.Join(dst, localizedPath), files)
	CheckFs(t, testDir.String(), fsExpected, fsActual)
}

func TestNestedRemoteRoots(t *testing.T) {
	fsExpected, fsActual, testDir := PrepareFs(t, nil, map[string]string{
		// TODO(annasong): Change the ref to the release after kustomize/v4.5.7.
		// We need changes to remote post-kustomize/v4.5.7.
		"kustomization.yaml": `resources:
- https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/remote?submodules=0&ref=master&timeout=300
`,
	})

	newDir := testDir.Join("dst")
	dst, err := localizer.Run(fsActual, testDir.String(), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	localizedPath, files := remotePathAndFiles(t)
	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, localizedPath),
	})
	SetupDir(t, fsExpected, filepath.Join(dst, localizedPath), files)
	CheckFs(t, testDir.String(), fsExpected, fsActual)
}

func TestResourcesRepoNotFile(t *testing.T) {
	const repo = "https://github.com/kubernetes-sigs/kustomize" + urlQuery
	kustomization := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, repo),
	}
	fsExpected, fsActual, testDir := PrepareFs(t, nil, kustomization)

	_, err := localizer.Run(fsActual, testDir.String(), "", testDir.Join("dst"))

	fileErr := fmt.Sprintf(`invalid resource at file "%s"`, repo)
	rootErr := fmt.Sprintf(`unable to localize root "%s": unable to find one of 'kustomization.yaml', 'kustomization.yml' or 'Kustomization'`, repo)
	var actualErr PathLocalizeError
	require.ErrorAs(t, err, &actualErr)
	require.Equal(t, repo, actualErr.Path)
	require.ErrorContains(t, actualErr.FileError, fileErr)
	require.ErrorContains(t, actualErr.RootError, rootErr)

	SetupDir(t, fsExpected, testDir.String(), kustomization)
	CheckFs(t, testDir.String(), fsExpected, fsActual)
}

func TestRemoteRootNoRef(t *testing.T) {
	const root = simpleURL + "?submodules=0&timeout=300"
	kustomization := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, root),
	}
	fsExpected, fsActual, testDir := PrepareFs(t, nil, kustomization)

	_, err := localizer.Run(fsActual, testDir.String(), "", testDir.Join("dst"))

	const fileErr = "invalid file reference: URL is a git repository"
	rootErr := fmt.Sprintf(`localize remote root "%s" missing ref query string parameter`, root)
	var actualErr PathLocalizeError
	require.ErrorAs(t, err, &actualErr)
	require.Equal(t, root, actualErr.Path)
	require.EqualError(t, actualErr.FileError, fileErr)
	require.EqualError(t, actualErr.RootError, rootErr)

	SetupDir(t, fsExpected, testDir.String(), kustomization)
	CheckFs(t, testDir.String(), fsExpected, fsActual)
}

func TestExistingCacheDir(t *testing.T) {
	const remoteFile = `https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/deployment.yaml`
	file := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`resources:
- %s
`, remoteFile),
		filepath.Join(LocalizeDir, "file"): "existing",
	}
	fsExpected, fsActual, testDir := PrepareFs(t, []string{LocalizeDir}, file)

	_, err := localizer.Run(fsActual, testDir.String(), "", testDir.Join("dst"))
	require.ErrorContains(t, err, fmt.Sprintf(`already contains localized-files needed to store file "%s"`, remoteFile))

	SetupDir(t, fsExpected, testDir.String(), file)
	CheckFs(t, testDir.String(), fsExpected, fsActual)
}

func TestHelmNestedHome(t *testing.T) {
	files := map[string]string{
		"kustomization.yaml": fmt.Sprintf(`helmGlobals:
  chartHome: %s
`, filepath.Join("nested", "dirs", "home")),
		filepath.Join("nested", "dirs", "home", "name", "values.yaml"): `
minecraftServer:
  difficulty: peaceful
`,
	}
	fsExpected, fsActual, testDir := PrepareFs(t, []string{
		filepath.Join("nested", "dirs", "home", "name"),
	}, files)

	newDir := testDir.Join("dst")
	dst, err := localizer.Run(fsActual, testDir.String(), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	SetupDir(t, fsExpected, dst, files)
	CheckFs(t, dst, fsExpected, fsActual)
}

func TestHelmLinkedHome(t *testing.T) {
	// scope
	// - target
	//   - kustomization
	//   - myValues.yaml
	//   - link to home
	// - home
	//   - name
	//     - values.yaml
	fsExpected, fsActual, scope := PrepareFs(t, []string{
		"target",
		filepath.Join("home", "name"),
	},
		map[string]string{
			filepath.Join("target", "Kustomization"): `helmCharts:
- name: name
  valuesFile: myValues.yaml
helmGlobals:
  chartHome: home-link
`,
			filepath.Join("target", "myValues.yaml"):     valuesFile,
			filepath.Join("home", "name", "values.yaml"): valuesFile,
		})
	link(t, scope, map[string]string{
		filepath.Join("target", "home-link"): "home",
	})

	newDir := scope.Join("dst")
	dst, err := localizer.Run(fsActual, scope.Join("target"), scope.String(), newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	SetupDir(t, fsExpected, dst, map[string]string{
		filepath.Join("target", "Kustomization"): fmt.Sprintf(`helmCharts:
- name: name
  valuesFile: myValues.yaml
helmGlobals:
  chartHome: %s
`, filepath.Join("..", "home")),
		filepath.Join("target", "myValues.yaml"):     valuesFile,
		filepath.Join("home", "name", "values.yaml"): valuesFile,
	})
	CheckFs(t, dst, fsExpected, fsActual)
}

func TestHelmLinkedDefaultHome(t *testing.T) {
	// target
	// - kustomization
	// - link to home (named charts)
	// - home
	//   - name
	//     - values.yaml
	fsExpected, fsActual, target := PrepareFs(t, []string{
		filepath.Join("home", "default"),
		filepath.Join("home", "same"),
	}, map[string]string{
		"kustomization.yaml": fmt.Sprintf(`helmCharts:
- name: default
helmChartInflationGenerator:
- chartHome: %s
  chartName: same
`, filepath.Join("home", "..", "charts")),
		filepath.Join("home", "default", "values.yaml"): valuesFile,
		filepath.Join("home", "same", "values.yaml"):    valuesFile,
	})
	link(t, target, map[string]string{"charts": "home"})

	newDir := target.Join("dst")
	dst, err := localizer.Run(fsActual, target.String(), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": `helmChartInflationGenerator:
- chartHome: charts
  chartName: same
helmCharts:
- name: default
`,
		filepath.Join("charts", "default", "values.yaml"): valuesFile,
		filepath.Join("charts", "same", "values.yaml"):    valuesFile,
	})
	CheckFs(t, dst, fsExpected, fsActual)
}

func TestHelmHomeEscapesScope(t *testing.T) {
	// test directory
	// - dir
	// - file
	// - target (and scope)
	//   - kustomization
	//   - home
	//     - link to dir
	//     - link to file
	fsExpected, fsActual, testDir := PrepareFs(t, []string{
		"dir",
		filepath.Join("target", "home"),
	}, map[string]string{
		"file": valuesFile,
		filepath.Join("target", "kustomization.yaml"): `helmGlobals:
  chartHome: home
`,
	})
	link(t, testDir, map[string]string{
		filepath.Join("target", "home", "dir-link"):  "dir",
		filepath.Join("target", "home", "file-link"): "file",
	})

	newDir := testDir.Join("dst")
	dst, err := localizer.Run(fsActual, testDir.Join("target"), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": `helmGlobals:
  chartHome: home
`,
	})
	require.NoError(t, fsExpected.Mkdir(filepath.Join(dst, "home")))
	CheckFs(t, dst, fsExpected, fsActual)
}

func TestSymlinkedFileSource(t *testing.T) {
	// target (and scope)
	// - kustomization
	// - file
	// - link to file
	fsExpected, fsActual, target := PrepareFs(t, nil, map[string]string{
		"kustomization.yaml": `configMapGenerator:
- files:
  - filename-used-as-key-in-configMap
`,
		"different-key": "properties",
	})
	link(t, target, map[string]string{
		"filename-used-as-key-in-configMap": "different-key",
	})

	newDir := target.Join("dst")
	dst, err := localizer.Run(fsActual, target.String(), "", newDir)
	require.NoError(t, err)
	require.Equal(t, newDir, dst)

	SetupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": `configMapGenerator:
- files:
  - different-key
`,
		"different-key": "properties",
	})
	CheckFs(t, dst, fsExpected, fsActual)
}
