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
)

func prepareFs(t *testing.T, files map[string]string) (memoryFs filesys.FileSystem, actualFs filesys.FileSystem, testDir filesys.ConfirmedDir) {
	t.Helper()

	memoryFs = filesys.MakeFsInMemory()
	actualFs = filesys.MakeFsOnDisk()

	testDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)

	setupDir(t, memoryFs, testDir.String(), files)
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

func getLocFilePath(t *testing.T, pathFromTestdata []string) string {
	t.Helper()

	localizedPathDirs := []string{LocalizeDir, "raw.githubusercontent.com", "kubernetes-sigs", "kustomize",
		"kustomize", "v4.5.7", "api", "krusty", "testdata"}
	return filepath.Join(append(localizedPathDirs, pathFromTestdata...)...)
}

// checkFs checks fsActual, the real file system, against fsExpected, a file system in memory, for contents
// in directory walkDir.
func checkFs(t *testing.T, walkDir string, fsExpected filesys.FileSystem, fsActual filesys.FileSystem) {
	t.Helper()

	err := fsExpected.Walk(walkDir, func(path string, info fs.FileInfo, err error) error {
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

	err = fsActual.Walk(walkDir, func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		// no symlinks yet
		require.NotEqual(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)
		require.True(t, fsExpected.Exists(path))
		return nil
	})
	require.NoError(t, err)
}

func TestRemoteFile(t *testing.T) {
	const kustf = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
openapi:
  path: %s
`
	fsExpected, fsActual, testDir := prepareFs(t, map[string]string{
		"kustomization.yaml": fmt.Sprintf(kustf, `https://raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/customschema.json`),
	})

	dst := testDir.Join("dst")
	err := localizer.Run(fsActual, testDir.String(), "", dst)
	require.NoError(t, err)

	localizedPath := getLocFilePath(t, []string{"customschema.json"})
	setupDir(t, fsExpected, dst, map[string]string{
		"kustomization.yaml": fmt.Sprintf(kustf, localizedPath),
		localizedPath:        customSchema,
	})
	checkFs(t, testDir.String(), fsExpected, fsActual)
}
