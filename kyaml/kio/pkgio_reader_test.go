// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "must specify package path")
	}
	assert.Nil(t, nodes)
}

func TestLocalPackageReader_Read_pkg(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("b_test.yaml"), readFileB)
	s.WriteFile(t, filepath.Join("c_test.yaml"), readFileC)
	s.WriteFile(t, filepath.Join("d_test.yaml"), readFileD)

	paths := []struct {
		path string
	}{
		{path: "./"},
		{path: s.Root},
	}
	for _, p := range paths {
		rfr := LocalPackageReader{PackagePath: p.path}
		nodes, err := rfr.Read()
		if !assert.NoError(t, err) {
			return
		}

		if !assert.Len(t, nodes, 5) {
			return
		}
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
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, expected[i], val) {
				return
			}
		}
	}
}

func TestLocalPackageReader_Read_JSON(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()

	s.WriteFile(t, filepath.Join("a_test.json"), []byte(`{
  "a": "b"
}`))
	s.WriteFile(t, filepath.Join("b_test.json"), []byte(`{
  "e": "f",
  "g": {
    "h": ["i", "j"]
  }
}`))

	paths := []struct {
		path string
	}{
		{path: "./"},
		{path: s.Root},
	}
	for _, p := range paths {
		rfr := LocalPackageReader{PackagePath: p.path, MatchFilesGlob: []string{"*.json"}}
		nodes, err := rfr.Read()
		if !assert.NoError(t, err) {
			return
		}

		if !assert.Len(t, nodes, 2) {
			t.FailNow()
		}
		// TODO: Fix https://github.com/go-yaml/yaml/issues/44 so these are printed correctly
		expected := []string{
			`{"a": "b", metadata: {annotations: {config.kubernetes.io/index: '0', config.kubernetes.io/path: 'a_test.json'}}}
`,
			`{"e": "f", "g": {"h": ["i", "j"]}, metadata: {annotations: {config.kubernetes.io/index: '0',
      config.kubernetes.io/path: 'b_test.json'}}}
`,
		}
		for i := range nodes {
			val, err := nodes[i].String()
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, expected[i], val) {
				return
			}
		}
	}
}

func TestLocalPackageReader_Read_file(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("b_test.yaml"), readFileB)

	paths := []struct {
		path string
	}{
		{path: "./"},
		{path: s.Root},
	}
	for _, p := range paths {
		rfr := LocalPackageReader{PackagePath: filepath.Join(p.path, "a_test.yaml")}
		nodes, err := rfr.Read()
		if !assert.NoError(t, err) {
			return
		}

		if !assert.Len(t, nodes, 2) {
			return
		}
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
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, expected[i], val) {
				return
			}
		}
	}
}

func TestLocalPackageReader_Read_pkgOmitAnnotations(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("b_test.yaml"), readFileB)

	paths := []struct {
		path string
	}{
		{path: "./"},
		{path: s.Root},
	}
	for _, p := range paths {
		// empty path
		rfr := LocalPackageReader{PackagePath: p.path, OmitReaderAnnotations: true}
		nodes, err := rfr.Read()
		if !assert.NoError(t, err) {
			return
		}

		if !assert.Len(t, nodes, 3) {
			return
		}
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
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, expected[i], val) {
				return
			}
		}
	}
}

func TestLocalPackageReader_Read_nestedDirs(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("a", "b", "b_test.yaml"), readFileB)

	paths := []struct {
		path string
	}{
		{path: "./"},
		{path: s.Root},
	}
	for _, p := range paths {
		// empty path
		rfr := LocalPackageReader{PackagePath: p.path}
		nodes, err := rfr.Read()
		if !assert.NoError(t, err) {
			assert.FailNow(t, err.Error())
		}

		if !assert.Len(t, nodes, 3) {
			return
		}
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
			if !assert.NoError(t, err) {
				return
			}
			want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
			if !assert.Equal(t, want, val) {
				return
			}
		}
	}
}

func TestLocalPackageReader_Read_matchRegex(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("a", "b", "b_test.yaml"), readFileB)

	// empty path
	rfr := LocalPackageReader{PackagePath: s.Root, MatchFilesGlob: []string{`a*.yaml`}}
	nodes, err := rfr.Read()
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	if !assert.Len(t, nodes, 2) {
		assert.FailNow(t, "wrong number items")
	}

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
		assert.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		assert.Equal(t, want, val)
	}
}

func TestLocalPackageReader_Read_skipSubpackage(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("a", "c", "c_test.yaml"), readFileB)
	s.WriteFile(t, filepath.Join("a", "c", "pkgFile"), pkgFile)

	// empty path
	rfr := LocalPackageReader{PackagePath: s.Root, PackageFileName: "pkgFile"}
	nodes, err := rfr.Read()
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	if !assert.Len(t, nodes, 2) {
		assert.FailNow(t, "wrong number items")
	}

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
		assert.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		assert.Equal(t, want, val)
	}
}

func TestLocalPackageReader_Read_includeSubpackage(t *testing.T) {
	s := SetupDirectories(t, filepath.Join("a", "b"), filepath.Join("a", "c"))
	defer s.Clean()
	s.WriteFile(t, filepath.Join("a", "b", "a_test.yaml"), readFileA)
	s.WriteFile(t, filepath.Join("a", "c", "c_test.yaml"), readFileB)
	s.WriteFile(t, filepath.Join("a", "c", "pkgFile"), pkgFile)

	// empty path
	rfr := LocalPackageReader{PackagePath: s.Root, IncludeSubpackages: true, PackageFileName: "pkgFile"}
	nodes, err := rfr.Read()
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	if !assert.Len(t, nodes, 3) {
		assert.FailNow(t, "wrong number items")
	}

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
		assert.NoError(t, err)
		want := strings.ReplaceAll(expected[i], "${SEP}", string(filepath.Separator))
		assert.Equal(t, want, val)
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
