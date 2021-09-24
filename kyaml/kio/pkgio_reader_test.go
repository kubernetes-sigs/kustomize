// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	. "sigs.k8s.io/kustomize/kyaml/kio"
)

var readFileA = []byte(`---
a: b #first
---
c: d # second
`)

var readFileB = []byte(`# second thing
e: f
g:
  h:
  - i # has a list
  - j
`)

var readFileC = []byte(`---
a: b #third
metadata:
  annotations:
`)

var readFileD = []byte(`---
a: b #forth
metadata:
`)

var readFileE = []byte(`---
a: b #first
---
- foo # second
- bar
`)

var pkgFile = []byte(``)

func TestLocalPackageReader_Read_empty(t *testing.T) {
	var r LocalPackageReader
	nodes, err := r.Read()
	require.Error(t, err)
	require.Contains(t, err.Error(), "must specify package path")
	require.Nil(t, nodes)
}

func TestLocalPackageReader_Read_pkg(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a_test.yaml", content: readFileA},
		{path: "b_test.yaml", content: readFileB},
		{path: "c_test.yaml", content: readFileC},
		{path: "d_test.yaml", content: readFileD},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath: path,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 5)
		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a_test.yaml'
`,
			`# second thing
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'b_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'b_test.yaml'
`,
			`a: b #third
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'c_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'c_test.yaml'
`,
			`a: b #forth
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'd_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'd_test.yaml'
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}

func TestLocalPackageReader_Read_pkgAndSkipFile(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a_test.yaml", content: readFileA},
		{path: "b_test.yaml", content: readFileB},
		{path: "c_test.yaml", content: readFileC},
		{path: "d_test.yaml", content: readFileD},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:  path,
			FileSkipFunc: func(relPath string) bool { return relPath == "d_test.yaml" },
			FileSystem:   filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 4)
		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a_test.yaml'
`,
			`# second thing
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'b_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'b_test.yaml'
`,
			`a: b #third
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'c_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'c_test.yaml'
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}

func TestLocalPackageReader_Read_JSON(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a_test.json", content: []byte(`{
			"a": "b"
		  }`),
		},
		{path: "b_test.json", content: []byte(`{
			"e": "f",
			"g": {
			  "h": ["i", "j"]
			}
		  }`),
		},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:    path,
			MatchFilesGlob: []string{"*.json"},
			FileSystem:     filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 2)
		expected := []string{
			`{"a": "b", metadata: {annotations: {config.kubernetes.io/index: '0', config.kubernetes.io/path: 'a_test.json', internal.config.kubernetes.io/index: '0', internal.config.kubernetes.io/path: 'a_test.json'}}}
`,
			`{"e": "f", "g": {"h": ["i", "j"]}, metadata: {annotations: {config.kubernetes.io/index: '0', config.kubernetes.io/path: 'b_test.json', internal.config.kubernetes.io/index: '0', internal.config.kubernetes.io/path: 'b_test.json'}}}
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}

func TestLocalPackageReader_Read_file(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a_test.yaml", content: readFileA},
		{path: "b_test.yaml", content: readFileB},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath: filepath.Join(path, "a_test.yaml"),
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 2)
		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a_test.yaml'
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}

func TestLocalPackageReader_Read_pkgOmitAnnotations(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a_test.yaml", content: readFileA},
		{path: "b_test.yaml", content: readFileB},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:           path,
			OmitReaderAnnotations: true,
			FileSystem:            filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 3)
		expected := []string{
			`a: b #first
`,
			`c: d # second
`,
			`# second thing
e: f
g:
  h:
  - i # has a list
  - j
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}

func TestLocalPackageReader_Read_PreserveSeqIndent(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a_test.yaml", content: readFileA},
		{path: "b_test.yaml", content: readFileB},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:       path,
			PreserveSeqIndent: true,
			FileSystem:        filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 3)
		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/seqindent: 'compact'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a_test.yaml'
    internal.config.kubernetes.io/seqindent: 'compact'
`,
			`# second thing
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'b_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'b_test.yaml'
    internal.config.kubernetes.io/seqindent: 'compact'
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}

func TestLocalPackageReader_Read_nestedDirs(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a/b/a_test.yaml", content: readFileA},
		{path: "a/b/b_test.yaml", content: readFileB},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath: path,
			FileSystem:  filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 3)
		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
			`# second thing
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a${SEP}b${SEP}b_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}b_test.yaml'
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
			require.Equal(t, want, val)
		}
	})
}

func TestLocalPackageReader_Read_matchRegex(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a/b/a_test.yaml", content: readFileA},
		{path: "a/b/b_test.yaml", content: readFileB},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:    path,
			MatchFilesGlob: []string{`a*.yaml`},
			FileSystem:     filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 2)

		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
		}

		for i, node := range nodes {
			val, err := node.String()
			require.NoError(t, err)
			want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
			require.Equal(t, want, val)
		}
	})
}

func TestLocalPackageReader_Read_skipSubpackage(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a/b/a_test.yaml", content: readFileA},
		{path: "a/c/c_test.yaml", content: readFileB},
		{path: "a/c/pkgFile", content: pkgFile},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:     path,
			PackageFileName: "pkgFile",
			FileSystem:      filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 2)

		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
		}

		for i, node := range nodes {
			val, err := node.String()
			require.NoError(t, err)
			want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
			require.Equal(t, want, val)
		}
	})
}

func TestLocalPackageReader_Read_includeSubpackage(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "a/b/a_test.yaml", content: readFileA},
		{path: "a/c/c_test.yaml", content: readFileB},
		{path: "a/c/pkgFile", content: pkgFile},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:        path,
			IncludeSubpackages: true,
			PackageFileName:    "pkgFile",
			FileSystem:         filesys.FileSystemOrOnDisk{FileSystem: mockFS},
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 3)

		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
			`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
			`# second thing
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'a${SEP}c${SEP}c_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'a${SEP}c${SEP}c_test.yaml'
`,
		}

		for i, node := range nodes {
			val, err := node.String()
			require.NoError(t, err)
			want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
			require.Equal(t, want, val)
		}
	})
}

type mockFile struct {
	path string
	// nil content implies this is a directory
	content []byte
}

func testOnDiskAndOnMem(t *testing.T, files []mockFile, f func(t *testing.T, path string, fs filesys.FileSystem)) {
	t.Run("on_disk", func(t *testing.T) {
		var dirs []string
		for _, file := range files {
			if file.content == nil {
				dirs = append(dirs, filepath.FromSlash(file.path))
			}
		}

		s := SetupDirectories(t, dirs...)
		defer s.Clean()
		for _, file := range files {
			if file.content != nil {
				s.WriteFile(t, filepath.FromSlash(file.path), file.content)
			}
		}

		f(t, "./", nil)
		f(t, s.Root, nil)
	})

	// TODO: Once fsnode supports Windows, we should also run the tests below.
	if runtime.GOOS == "windows" {
		return
	}

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		for _, file := range files {
			path := filepath.FromSlash(file.path)
			if file.content == nil {
				require.NoError(t, fs.MkdirAll(path))
			} else {
				require.NoError(t, fs.WriteFile(path, file.content))
			}
		}

		f(t, "/", fs)
	})
}

func TestLocalPackageReader_ReadBareSeqNodes(t *testing.T) {
	testOnDiskAndOnMem(t, []mockFile{
		{path: "a/b"},
		{path: "a/c"},
		{path: "e_test.yaml", content: readFileE},
	}, func(t *testing.T, path string, mockFS filesys.FileSystem) {
		rfr := LocalPackageReader{
			PackagePath:     path,
			FileSystem:      filesys.FileSystemOrOnDisk{FileSystem: mockFS},
			WrapBareSeqNode: true,
		}
		nodes, err := rfr.Read()
		require.NoError(t, err)
		require.Len(t, nodes, 2)
		expected := []string{
			`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'e_test.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'e_test.yaml'
`,
			`bareSeqNodeWrappingKey:
- foo # second
- bar
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'e_test.yaml'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'e_test.yaml'
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			require.NoError(t, err)
			require.Equal(t, expected[i], val)
		}
	})
}
