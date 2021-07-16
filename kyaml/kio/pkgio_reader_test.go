// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"path/filepath"
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

var pkgFile = []byte(``)

func TestLocalPackageReader_Read_empty(t *testing.T) {
	var r LocalPackageReader
	nodes, err := r.Read()
	require.Error(t, err)
	require.Contains(t, err.Error(), "must specify package path")
	require.Nil(t, nodes)
}

func TestLocalPackageReader_Read_pkg(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, "a_test.yaml", readFileA)
		s.WriteFile(t, "b_test.yaml", readFileB)
		s.WriteFile(t, "c_test.yaml", readFileC)
		s.WriteFile(t, "d_test.yaml", readFileD)

		testLocalPackageReaderReadPkg(t, "./", nil)
		testLocalPackageReaderReadPkg(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile("a_test.yaml", readFileA))
		require.NoError(t, fs.WriteFile("b_test.yaml", readFileB))
		require.NoError(t, fs.WriteFile("c_test.yaml", readFileC))
		require.NoError(t, fs.WriteFile("d_test.yaml", readFileD))

		testLocalPackageReaderReadPkg(t, "/", fs)
	})
}

func testLocalPackageReaderReadPkg(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
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
`,
		`a: b #third
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'c_test.yaml'
`,
		`a: b #forth
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'd_test.yaml'
`,
	}
	for i := range nodes {
		val, err := nodes[i].String()
		require.NoError(t, err)
		require.Equal(t, expected[i], val)
	}
}

func TestLocalPackageReader_Read_pkgAndSkipFile(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, "a_test.yaml", readFileA)
		s.WriteFile(t, "b_test.yaml", readFileB)
		s.WriteFile(t, "c_test.yaml", readFileC)
		s.WriteFile(t, "d_test.yaml", readFileD)

		testLocalPackageReaderReadPkgAndSkipFile(t, "./", nil)
		testLocalPackageReaderReadPkgAndSkipFile(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile("a_test.yaml", readFileA))
		require.NoError(t, fs.WriteFile("b_test.yaml", readFileB))
		require.NoError(t, fs.WriteFile("c_test.yaml", readFileC))
		require.NoError(t, fs.WriteFile("d_test.yaml", readFileD))

		testLocalPackageReaderReadPkgAndSkipFile(t, "/", fs)
	})
}

func testLocalPackageReaderReadPkgAndSkipFile(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
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
`,
		`a: b #third
metadata:
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'c_test.yaml'
`,
	}
	for i := range nodes {
		val, err := nodes[i].String()
		require.NoError(t, err)
		require.Equal(t, expected[i], val)
	}
}

func TestLocalPackageReader_Read_JSON(t *testing.T) {
	aTestJSON := []byte(`{
		"a": "b"
	  }`)
	bTestJSON := []byte(`{
		"e": "f",
		"g": {
		  "h": ["i", "j"]
		}
	  }`)

	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()

		s.WriteFile(t, "a_test.json", aTestJSON)
		s.WriteFile(t, "b_test.json", bTestJSON)

		testLocalPackageReaderReadJSON(t, "./", nil)
		testLocalPackageReaderReadJSON(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile("a_test.json", aTestJSON))
		require.NoError(t, fs.WriteFile("b_test.json", bTestJSON))

		testLocalPackageReaderReadJSON(t, "/", fs)
	})
}

func testLocalPackageReaderReadJSON(t *testing.T, path string, mockFS filesys.FileSystem) {
	rfr := LocalPackageReader{
		PackagePath:    path,
		MatchFilesGlob: []string{"*.json"},
	}
	rfr.FileSystem.Set(mockFS)
	nodes, err := rfr.Read()
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	expected := []string{
		`{"a": "b", metadata: {annotations: {config.kubernetes.io/index: '0', config.kubernetes.io/path: 'a_test.json'}}}
`,
		`{"e": "f", "g": {"h": ["i", "j"]}, metadata: {annotations: {config.kubernetes.io/index: '0', config.kubernetes.io/path: 'b_test.json'}}}
`,
	}
	for i := range nodes {
		val, err := nodes[i].String()
		require.NoError(t, err)
		require.Equal(t, expected[i], val)
	}
}

func TestLocalPackageReader_Read_file(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("b_test.yaml"), readFileB)

		testLocalPackageReaderReadFile(t, "./", nil)
		testLocalPackageReaderReadFile(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile("a_test.yaml", readFileA))
		require.NoError(t, fs.WriteFile("b_test.yaml", readFileB))

		testLocalPackageReaderReadFile(t, "/", fs)
	})
}

func testLocalPackageReaderReadFile(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
`,
	}
	for i := range nodes {
		val, err := nodes[i].String()
		require.NoError(t, err)
		require.Equal(t, expected[i], val)
	}
}

func TestLocalPackageReader_Read_pkgOmitAnnotations(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("b_test.yaml"), readFileB)

		testLocalPackageReaderReadPkgOmitAnnotations(t, "./", nil)
		testLocalPackageReaderReadPkgOmitAnnotations(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile("a_test.yaml", readFileA))
		require.NoError(t, fs.WriteFile("b_test.yaml", readFileB))

		testLocalPackageReaderReadPkgOmitAnnotations(t, "/", fs)
	})
}

func testLocalPackageReaderReadPkgOmitAnnotations(t *testing.T, path string, mockFS filesys.FileSystem) {
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
}

func TestLocalPackageReader_Read_PreserveSeqIndent(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("b_test.yaml"), readFileB)

		testLocalPackageReaderReadPreserveSeqIndent(t, "./", nil)
		testLocalPackageReaderReadPreserveSeqIndent(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile("a_test.yaml", readFileA))
		require.NoError(t, fs.WriteFile("b_test.yaml", readFileB))

		testLocalPackageReaderReadPreserveSeqIndent(t, "/", fs)
	})
}

func testLocalPackageReaderReadPreserveSeqIndent(t *testing.T, path string, mockFS filesys.FileSystem) {
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
    internal.config.kubernetes.io/seqindent: 'compact'
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a_test.yaml'
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
    internal.config.kubernetes.io/seqindent: 'compact'
`,
	}
	for i := range nodes {
		val, err := nodes[i].String()
		require.NoError(t, err)
		require.Equal(t, expected[i], val)
	}
}

func TestLocalPackageReader_Read_nestedDirs(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("a", "b", "b_test.yaml"), readFileB)

		testLocalPackageReaderReadNestedDirs(t, "./", nil)
		testLocalPackageReaderReadNestedDirs(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "b", "a_test.yaml"), readFileA))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "b", "b_test.yaml"), readFileB))

		testLocalPackageReaderReadNestedDirs(t, "/", fs)
	})
}

func testLocalPackageReaderReadNestedDirs(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
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
`,
	}
	for i := range nodes {
		val, err := nodes[i].String()
		require.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		require.Equal(t, want, val)
	}
}

func TestLocalPackageReader_Read_matchRegex(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("a", "b", "b_test.yaml"), readFileB)

		testLocalPackageReaderReadMatchRegex(t, "./", nil)
		testLocalPackageReaderReadMatchRegex(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "b", "a_test.yaml"), readFileA))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "b", "b_test.yaml"), readFileB))

		testLocalPackageReaderReadMatchRegex(t, "/", fs)
	})
}

func testLocalPackageReaderReadMatchRegex(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
	}

	for i, node := range nodes {
		val, err := node.String()
		require.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		require.Equal(t, want, val)
	}
}

func TestLocalPackageReader_Read_skipSubpackage(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("a", "c", "c_test.yaml"), readFileB)
		s.WriteFile(t, filepath.Join("a", "c", "pkgFile"), pkgFile)

		testLocalPackageReaderReadSkipSubpackage(t, "./", nil)
		testLocalPackageReaderReadSkipSubpackage(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "b", "a_test.yaml"), readFileA))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "c", "c_test.yaml"), readFileB))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "c", "pkgFile"), pkgFile))

		testLocalPackageReaderReadSkipSubpackage(t, "/", fs)
	})
}

func testLocalPackageReaderReadSkipSubpackage(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
`,
	}

	for i, node := range nodes {
		val, err := node.String()
		require.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		require.Equal(t, want, val)
	}
}

func TestLocalPackageReader_Read_includeSubpackage(t *testing.T) {
	t.Run("on_disk", func(t *testing.T) {
		s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
		defer s.Clean()
		s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
		s.WriteFile(t, filepath.Join("a", "c", "c_test.yaml"), readFileB)
		s.WriteFile(t, filepath.Join("a", "c", "pkgFile"), pkgFile)

		testLocalPackageReaderReadIncludeSubpackage(t, "./", nil)
		testLocalPackageReaderReadIncludeSubpackage(t, s.Root, nil)
	})

	t.Run("on_mem", func(t *testing.T) {
		fs := filesys.MakeFsInMemory()
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "b")))
		require.NoError(t, fs.MkdirAll(filepath.Join("a", "c")))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "b", "a_test.yaml"), readFileA))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "c", "c_test.yaml"), readFileB))
		require.NoError(t, fs.WriteFile(filepath.Join("a", "c", "pkgFile"), pkgFile))

		testLocalPackageReaderReadIncludeSubpackage(t, "/", fs)
	})
}

func testLocalPackageReaderReadIncludeSubpackage(t *testing.T, path string, mockFS filesys.FileSystem) {
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
`,
		`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'a${SEP}b${SEP}a_test.yaml'
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
`,
	}

	for i, node := range nodes {
		val, err := node.String()
		require.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		require.Equal(t, want, val)
	}
}

// func TestLocalPackageReaderWriter_DeleteFiles(t *testing.T) {
// 	g, _, clean := testutil.SetupDefaultRepoAndWorkspace(t)
// 	defer clean()
// 	if !assert.NoError(t, os.Chdir(g.RepoDirectory)) {
// 		return
// 	}
//
// 	rw := LocalPackageReadWriter{PackagePath: "."}
// 	nodes, err := rw.Read()
// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}
// 	_, err = os.Stat(filepath.Join("java", "java-deployment.resource.yaml"))
// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}
//
// 	// delete one of the nodes
// 	var newNodes []*yaml.RNode
// 	for i := range nodes {
// 		meta, err := nodes[i].GetMeta()
// 		if !assert.NoError(t, err) {
// 			t.FailNow()
// 		}
// 		if meta.Name == "app" && meta.Kind == "Deployment" {
// 			continue
// 		}
// 		newNodes = append(newNodes, nodes[i])
// 	}
//
// 	if !assert.NoError(t, rw.Write(newNodes)) {
// 		t.FailNow()
// 	}
//
// 	_, err = os.Stat(filepath.Join("java", "java-deployment.resource.yaml"))
// 	if !assert.Error(t, err) {
// 		t.FailNow()
// 	}
//
// 	diff, err := copyutil.Diff(filepath.Join(g.DatasetDirectory, testutil.Dataset1), ".")
// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}
//
// 	assert.ElementsMatch(t,
// 		diff.List(),
// 		[]string{filepath.Join("java", "java-deployment.resource.yaml")})
// }
