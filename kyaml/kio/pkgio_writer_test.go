// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// TestLocalPackageWriter_Write tests:
// - ReaderAnnotations are cleared when writing the Resources
func TestLocalPackageWriter_Write(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	w := LocalPackageWriter{PackagePath: d}
	err := w.Write([]*yaml.RNode{node2, node1, node3})
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	b, err := ioutil.ReadFile(filepath.Join(d, "a", "b", "a_test.yaml"))
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, `a: b #first
---
c: d # second
`, string(b))

	b, err = ioutil.ReadFile(filepath.Join(d, "a", "b", "b_test.yaml"))
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
`, string(b))
}

// TestLocalPackageWriter_Write_keepReaderAnnotations tests:
// - ReaderAnnotations are kept when writing the Resources
func TestLocalPackageWriter_Write_keepReaderAnnotations(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	w := LocalPackageWriter{PackagePath: d, KeepReaderAnnotations: true}
	err := w.Write([]*yaml.RNode{node2, node1, node3})
	if !assert.NoError(t, err) {
		return
	}

	b, err := ioutil.ReadFile(filepath.Join(d, "a", "b", "a_test.yaml"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: a/b/a_test.yaml
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: a/b/a_test.yaml
`, string(b))

	b, err = ioutil.ReadFile(filepath.Join(d, "a", "b", "b_test.yaml"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: a/b/b_test.yaml
`, string(b))
}

// TestLocalPackageWriter_Write_clearAnnotations tests:
// - ClearAnnotations are removed from Resources
func TestLocalPackageWriter_Write_clearAnnotations(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	w := LocalPackageWriter{PackagePath: d, ClearAnnotations: []string{"config.kubernetes.io/mode"}}
	err := w.Write([]*yaml.RNode{node2, node1, node3})
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	b, err := ioutil.ReadFile(filepath.Join(d, "a", "b", "a_test.yaml"))
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, `a: b #first
---
c: d # second
`, string(b))

	b, err = ioutil.ReadFile(filepath.Join(d, "a", "b", "b_test.yaml"))
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
`, string(b))
}

// TestLocalPackageWriter_Write_failRelativePath tests:
// - If a relative path above the package is defined, write fails
func TestLocalPackageWriter_Write_failRelativePath(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

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
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	w := LocalPackageWriter{PackagePath: d}
	err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "resource must be written under package")
	}
}

// TestLocalPackageWriter_Write_invalidIndex tests:
// - If a non-int index is given, fail
func TestLocalPackageWriter_Write_invalidIndex(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

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
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	w := LocalPackageWriter{PackagePath: d}
	err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unable to parse config.kubernetes.io/index")
	}
}

// TestLocalPackageWriter_Write_absPath tests:
// - If config.kubernetes.io/path is absolute, fail
func TestLocalPackageWriter_Write_absPath(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	node4, err := yaml.Parse(fmt.Sprintf(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: a
    config.kubernetes.io/path: "%s/a/b/b_test.yaml" # use a different path, should still collide
`, d))
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	w := LocalPackageWriter{PackagePath: d}
	err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "package paths may not be absolute paths")
	}
}

// TestLocalPackageWriter_Write_missingIndex tests:
// - If config.kubernetes.io/path is missing, fail
func TestLocalPackageWriter_Write_missingPath(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	node4, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: a
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	w := LocalPackageWriter{PackagePath: d}
	err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "config.kubernetes.io/path")
	}
}

// TestLocalPackageWriter_Write_missingIndex tests:
// - If config.kubernetes.io/index is missing, fail
func TestLocalPackageWriter_Write_missingIndex(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	node4, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/path: a/a.yaml
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	w := LocalPackageWriter{PackagePath: d}
	err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "config.kubernetes.io/index")
	}
}

// TestLocalPackageWriter_Write_pathIsDir tests:
// - If  config.kubernetes.io/path is a directory, fail
func TestLocalPackageWriter_Write_pathIsDir(t *testing.T) {
	d, node1, node2, node3 := getWriterInputs(t)
	defer os.RemoveAll(d)

	node4, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/path: a/
    config.kubernetes.io/index: 0
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	w := LocalPackageWriter{PackagePath: d}
	err = w.Write([]*yaml.RNode{node2, node1, node3, node4})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "config.kubernetes.io/path cannot be a directory")
	}
}

func getWriterInputs(t *testing.T) (string, *yaml.RNode, *yaml.RNode, *yaml.RNode) {
	node1, err := yaml.Parse(`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	node2, err := yaml.Parse(`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	node3, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/b_test.yaml"
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	d, err := ioutil.TempDir("", "kyaml-test")
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	if !assert.NoError(t, os.MkdirAll(filepath.Join(d, "a"), 0700)) {
		assert.FailNow(t, "")
	}
	return d, node1, node2, node3
}
