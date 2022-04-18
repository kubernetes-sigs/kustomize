// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestIgnoreFilesMatcher_readIgnoreFile(t *testing.T) {
	testCases := []struct {
		name            string
		writeIgnoreFile bool
		isMatch         bool
	}{
		{
			name:            "has .krmignore file",
			writeIgnoreFile: true,
			isMatch:         true,
		},
		{
			name:            "no .krmignore file",
			writeIgnoreFile: false,
			isMatch:         false,
		},
	}

	const (
		ignoreFileName = ".krmignore"
		testFileName   = "testfile.yaml"
		ignoreFileBody = "\n" + testFileName + "\n"
	)

	fsMakers := map[string]func(bool) (string, filesys.FileSystem){
		// onDisk creates a temp directory and returns a nil FileSystem, testing
		// the normal conditions under which ignoreFileMatcher is used.
		"onDisk": func(writeIgnoreFile bool) (string, filesys.FileSystem) { //nolint:unparam
			dir := t.TempDir()

			if writeIgnoreFile {
				ignoreFilePath := filepath.Join(dir, ignoreFileName)
				require.NoError(t, ioutil.WriteFile(ignoreFilePath, []byte(ignoreFileBody), 0600))
			}
			testFilePath := filepath.Join(dir, testFileName)
			require.NoError(t, ioutil.WriteFile(testFilePath, []byte{}, 0600))
			return dir, nil
		},

		// inMem creates an in-memory FileSystem and returns it.
		"inMem": func(writeIgnoreFile bool) (string, filesys.FileSystem) {
			fs := filesys.MakeEmptyDirInMemory()
			if writeIgnoreFile {
				require.NoError(t, fs.WriteFile(ignoreFileName, []byte(ignoreFileBody)))
			}
			require.NoError(t, fs.WriteFile(testFileName, nil))
			return ".", fs
		},
	}

	for name, fsMaker := range fsMakers {
		t.Run(name, func(t *testing.T) {
			fsMaker := fsMaker
			for i := range testCases {
				test := testCases[i]
				dir, fs := fsMaker(test.writeIgnoreFile)
				t.Run(test.name, func(t *testing.T) {
					m := ignoreFilesMatcher{}
					m.fs.Set(fs)
					require.NoError(t, m.readIgnoreFile(dir))
					require.Equal(t, test.isMatch, m.matchFile(filepath.Join(dir, testFileName)))
				})
			}
		})
	}
}

var (
	readFileA = []byte(`
a: a
---
c: c
`)
	readFileB = []byte(`
b: b
`)
)

func TestLocalPackageReader_Read_ignoreFile(t *testing.T) {
	testCases := []struct {
		name        string
		directories []string
		files       map[string][]byte
		expected    []string
	}{
		{
			name: "ignore file",
			directories: []string{
				filepath.Join("a", "b"),
				filepath.Join("a", "c"),
			},
			files: map[string][]byte{
				"pkgFile":                              {},
				filepath.Join("a", "b", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				".krmignore": []byte(`
a/c/c_test.yaml
`,
				),
			},
			expected: []string{
				`a: a`,
				`c: c`,
			},
		},
		{
			name: "ignore folder",
			directories: []string{
				filepath.Join("a", "b"),
				filepath.Join("a", "c"),
			},
			files: map[string][]byte{
				"pkgFile":                              {},
				filepath.Join("a", "b", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				".krmignore": []byte(`
a/c
`,
				),
			},
			expected: []string{
				`a: a`,
				`c: c`,
			},
		},
		{
			name: "krmignore file in subpackage",
			directories: []string{
				filepath.Join("a", "c"),
			},
			files: map[string][]byte{
				"pkgFile":                              {},
				filepath.Join("a", "c", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				".krmignore": []byte(`
d/e/f.yaml
`,
				),
				filepath.Join("a", "c", "pkgFile"): {},
				filepath.Join("a", "c", ".krmignore"): []byte(`
a_test.yaml
`),
			},
			expected: []string{
				`b: b`,
			},
		},
		{
			name: "krmignore files does not affect subpackages",
			directories: []string{
				filepath.Join("a", "c"),
			},
			files: map[string][]byte{
				"pkgFile":                              {},
				filepath.Join("a", "c", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				".krmignore": []byte(`
a/c/c_test.yaml
`,
				),
				filepath.Join("a", "c", "pkgFile"): {},
				filepath.Join("a", "c", ".krmignore"): []byte(`
a_test.yaml
`),
			},
			expected: []string{
				`b: b`,
			},
		},
		{
			name: "handles a combination of packages and directories",
			directories: []string{
				"a",
				filepath.Join("d", "e"),
				"f",
			},
			files: map[string][]byte{
				"pkgFile":                                {},
				filepath.Join("d", "pkgFile"):            {},
				filepath.Join("d", "e", "pkgFile"):       {},
				filepath.Join("f", "pkgFile"):            {},
				"manifest.yaml":                          []byte(`root: root`),
				filepath.Join("a", "manifest.yaml"):      []byte(`a: a`),
				filepath.Join("d", "manifest.yaml"):      []byte(`d: d`),
				filepath.Join("d", "e", "manifest.yaml"): []byte(`e: e`),
				filepath.Join("f", "manifest.yaml"):      []byte(`f: f`),
				filepath.Join("d", ".krmignore"): []byte(`
manifest.yaml
`),
			},
			expected: []string{
				`a: a`,
				`e: e`,
				`f: f`,
				`root: root`,
			},
		},
		{
			name: "ignore file can exclude subpackages",
			directories: []string{
				"a",
			},
			files: map[string][]byte{
				"pkgFile":                           {},
				filepath.Join("a", "pkgFile"):       {},
				"manifest.yaml":                     []byte(`root: root`),
				filepath.Join("a", "manifest.yaml"): []byte(`a: a`),
				".krmignore": []byte(`
a
`),
			},
			expected: []string{
				`root: root`,
			},
		},
	}

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			s := SetupDirectories(t, test.directories...)
			defer s.Clean()
			for path, content := range test.files {
				s.WriteFile(t, path, content)
			}

			// empty path
			rfr := LocalPackageReader{
				PackagePath:           s.Root,
				IncludeSubpackages:    true,
				PackageFileName:       "pkgFile",
				OmitReaderAnnotations: true,
			}
			nodes, err := rfr.Read()
			if !assert.NoError(t, err) {
				assert.FailNow(t, err.Error())
			}

			if !assert.Len(t, nodes, len(test.expected)) {
				assert.FailNow(t, "wrong number items")
			}

			for i, node := range nodes {
				val, err := node.String()
				assert.NoError(t, err)
				want := strings.ReplaceAll(test.expected[i], "${SEP}", string(filepath.Separator))
				assert.Equal(t, strings.TrimSpace(want), strings.TrimSpace(val))
			}
		})
	}
}
