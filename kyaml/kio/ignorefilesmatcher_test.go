// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "kyaml-test")
			if !assert.NoError(t, err) {
				assert.FailNow(t, err.Error())
			}

			if test.writeIgnoreFile {
				ignoreFilePath := filepath.Join(dir, ".krmignore")
				err = ioutil.WriteFile(ignoreFilePath, []byte(`
testfile.yaml
`), 0600)
				if !assert.NoError(t, err) {
					assert.FailNow(t, err.Error())
				}
			}
			testFilePath := filepath.Join(dir, "testfile.yaml")
			err = ioutil.WriteFile(testFilePath, []byte{}, 0600)
			if !assert.NoError(t, err) {
				assert.FailNow(t, err.Error())
			}

			ignoreFilesMatcher := ignoreFilesMatcher{}
			err = ignoreFilesMatcher.readIgnoreFile(dir)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			assert.Equal(t, test.isMatch, ignoreFilesMatcher.matchFile(testFilePath))
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
				filepath.Join("pkgFile"):               {},
				filepath.Join("a", "b", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				filepath.Join(".krmignore"): []byte(`
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
				filepath.Join("pkgFile"):               {},
				filepath.Join("a", "b", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				filepath.Join(".krmignore"): []byte(`
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
				filepath.Join("pkgFile"):               {},
				filepath.Join("a", "c", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				filepath.Join(".krmignore"): []byte(`
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
				filepath.Join("pkgFile"):               {},
				filepath.Join("a", "c", "a_test.yaml"): readFileA,
				filepath.Join("a", "c", "c_test.yaml"): readFileB,
				filepath.Join(".krmignore"): []byte(`
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
				filepath.Join("a"),
				filepath.Join("d", "e"),
				filepath.Join("f"),
			},
			files: map[string][]byte{
				filepath.Join("pkgFile"):                 {},
				filepath.Join("d", "pkgFile"):            {},
				filepath.Join("d", "e", "pkgFile"):       {},
				filepath.Join("f", "pkgFile"):            {},
				filepath.Join("manifest.yaml"):           []byte(`root: root`),
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
				filepath.Join("a"),
			},
			files: map[string][]byte{
				filepath.Join("pkgFile"):            {},
				filepath.Join("a", "pkgFile"):       {},
				filepath.Join("manifest.yaml"):      []byte(`root: root`),
				filepath.Join("a", "manifest.yaml"): []byte(`a: a`),
				filepath.Join(".krmignore"): []byte(`
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
