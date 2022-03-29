// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// TestLocalPackageWriter_Write tests:
// - ReaderAnnotations are cleared when writing the Resources
func TestLocalPackageWriter_Write(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		w := LocalPackageWriter{
			PackagePath: d,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err := w.Write([]*yaml.RNode{node2, node1, node3})
		require.NoError(t, err)

		b, err := fs.ReadFile(filepath.Join(d, "a", "b", "a_test.yaml"))
		require.NoError(t, err)
		require.Equal(t, `a: b #first
---
c: d # second
`, string(b))

		b, err = fs.ReadFile(filepath.Join(d, "a", "b", "b_test.yaml"))
		require.NoError(t, err)
		require.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
`, string(b))
	})
}

// TestLocalPackageWriter_Write_keepReaderAnnotations tests:
// - ReaderAnnotations are kept when writing the Resources
func TestLocalPackageWriter_Write_keepReaderAnnotations(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		w := LocalPackageWriter{
			PackagePath:           d,
			KeepReaderAnnotations: true,
			FileSystem:            filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err := w.Write([]*yaml.RNode{node2, node1, node3})
		require.NoError(t, err)

		b, err := fs.ReadFile(filepath.Join(d, "a", "b", "a_test.yaml"))
		require.NoError(t, err)
		require.Equal(t, `a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: "0"
    config.kubernetes.io/path: "a/b/a_test.yaml"
    internal.config.kubernetes.io/path: 'a/b/a_test.yaml'
    internal.config.kubernetes.io/index: '0'
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: "1"
    config.kubernetes.io/path: "a/b/a_test.yaml"
    internal.config.kubernetes.io/path: 'a/b/a_test.yaml'
    internal.config.kubernetes.io/index: '1'
`, string(b))

		b, err = fs.ReadFile(filepath.Join(d, "a", "b", "b_test.yaml"))
		require.NoError(t, err)
		require.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: "0"
    config.kubernetes.io/path: "a/b/b_test.yaml"
    internal.config.kubernetes.io/path: 'a/b/b_test.yaml'
    internal.config.kubernetes.io/index: '0'
`, string(b))
	})
}

// TestLocalPackageWriter_Write_clearAnnotations tests:
// - ClearAnnotations are removed from Resources
func TestLocalPackageWriter_Write_clearAnnotations(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		w := LocalPackageWriter{
			PackagePath:      d,
			ClearAnnotations: []string{"config.kubernetes.io/mode"},
			FileSystem:       filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err := w.Write([]*yaml.RNode{node2, node1, node3})
		require.NoError(t, err)

		b, err := fs.ReadFile(filepath.Join(d, "a", "b", "a_test.yaml"))
		require.NoError(t, err)
		require.Equal(t, `a: b #first
---
c: d # second
`, string(b))

		b, err = fs.ReadFile(filepath.Join(d, "a", "b", "b_test.yaml"))
		require.NoError(t, err)
		require.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
`, string(b))
	})
}

// TestLocalPackageWriter_Write_failRelativePath tests:
// - If a relative path above the package is defined, write fails
func TestLocalPackageWriter_Write_failRelativePath(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		node4, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/../../../b_test.yaml"
`)
		require.NoError(t, err)

		w := LocalPackageWriter{
			PackagePath: d,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "resource must be written under package")
		}
	})
}

// TestLocalPackageWriter_Write_invalidIndex tests:
// - If a non-int index is given, fail
func TestLocalPackageWriter_Write_invalidIndex(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		node4, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: a
    config.kubernetes.io/path: "a/b/b_test.yaml" # use a different path, should still collide
`)
		require.NoError(t, err)

		w := LocalPackageWriter{
			PackagePath: d,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "unable to parse config.kubernetes.io/index")
		}
	})
}

// TestLocalPackageWriter_Write_absPath tests:
// - If config.kubernetes.io/path is absolute, fail
func TestLocalPackageWriter_Write_absPath(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		d = filepath.ToSlash(d)

		n4 := fmt.Sprintf(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: a
    config.kubernetes.io/path: "%s/a/b/b_test.yaml" # use a different path, should still collide
`, d)
		node4, err := yaml.Parse(n4)
		testutil.AssertNoError(t, err, n4)

		w := LocalPackageWriter{
			PackagePath: d,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
		testutil.AssertErrorContains(t, err, "package paths may not be absolute paths")
	})
}

// TestLocalPackageWriter_Write_missingPath tests:
// - If config.kubernetes.io/path or index are missing, then default them
func TestLocalPackageWriter_Write_missingAnnotations(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		node4String := `e: f
g:
  h:
  - i # has a list
  - j
kind: Foo
metadata:
  name: bar
`
		node4, err := yaml.Parse(node4String)
		require.NoError(t, err)

		w := LocalPackageWriter{
			PackagePath: d,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
		require.NoError(t, err)
		b, err := fs.ReadFile(filepath.Join(d, "foo_bar.yaml"))
		require.NoError(t, err)
		require.Equal(t, node4String, string(b))
	})
}

// TestLocalPackageWriter_Write_pathIsDir tests:
// - If  config.kubernetes.io/path is a directory, fail
func TestLocalPackageWriter_Write_pathIsDir(t *testing.T) {
	testWriterOnDiskAndOnMem(t, func(t *testing.T, fs filesys.FileSystem) {
		d, node1, node2, node3, cleanup := getWriterInputs(t, fs)
		defer cleanup()

		node4, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/path: a/
    config.kubernetes.io/index: "0"
`)
		require.NoError(t, err)

		w := LocalPackageWriter{
			PackagePath: d,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: fs},
		}
		err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
		require.Error(t, err)
		require.Contains(t, err.Error(), "config.kubernetes.io/path cannot be a directory")
	})
}

func testWriterOnDiskAndOnMem(t *testing.T, f func(t *testing.T, fs filesys.FileSystem)) {
	t.Run("on_disk", func(t *testing.T) { f(t, filesys.MakeFsOnDisk()) })
	// TODO: Once fsnode supports Windows, these tests should also be run.
	if runtime.GOOS != "windows" {
		t.Run("on_mem", func(t *testing.T) { f(t, filesys.MakeFsInMemory()) })
	}
}

func getWriterInputs(t *testing.T, mockFS filesys.FileSystem) (string, *yaml.RNode, *yaml.RNode, *yaml.RNode, func()) {
	node1, err := yaml.Parse(`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: "0"
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	require.NoError(t, err)
	node2, err := yaml.Parse(`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: "1"
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	require.NoError(t, err)
	node3, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: "0"
    config.kubernetes.io/path: "a/b/b_test.yaml"
`)
	require.NoError(t, err)

	// These two lines are similar to calling ioutil.TempDir, but we don't actually create any directory.
	rand.Seed(time.Now().Unix())
	path := filepath.Join(os.TempDir(), fmt.Sprintf("kyaml-test%d", rand.Int31())) //nolint:gosec
	require.NoError(t, mockFS.MkdirAll(filepath.Join(path, "a")))
	return path, node1, node2, node3, func() { require.NoError(t, mockFS.RemoveAll(path)) }
}
