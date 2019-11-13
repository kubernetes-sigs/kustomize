// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// TestByteWriter_Write_withoutAnnotations tests:
// - Resource Config ordering is preserved if no annotations are present
func TestByteWriter_Write_wrapped(t *testing.T) {
	node1, err := yaml.Parse(`a: b #first
`)
	if !assert.NoError(t, err) {
		return
	}
	node2, err := yaml.Parse(`c: d # second
`)
	if !assert.NoError(t, err) {
		return
	}
	node3, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
`)
	if !assert.NoError(t, err) {
		return
	}

	buff := &bytes.Buffer{}
	err = ByteWriter{
		Sort:               true,
		Writer:             buff,
		FunctionConfig:     node3,
		WrappingKind:       ResourceListKind,
		WrappingApiVersion: ResourceListApiVersion}.
		Write([]*yaml.RNode{node2, node1})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- c: d # second
- a: b #first
functionConfig:
  e: f
  g:
    h:
    - i # has a list
    - j
`, buff.String())
}

// TestByteWriter_Write_withoutAnnotations tests:
// - Resource Config ordering is preserved if no annotations are present
func TestByteWriter_Write_withoutAnnotations(t *testing.T) {
	node1, err := yaml.Parse(`a: b #first
`)
	if !assert.NoError(t, err) {
		return
	}
	node2, err := yaml.Parse(`c: d # second
`)
	if !assert.NoError(t, err) {
		return
	}
	node3, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
`)
	if !assert.NoError(t, err) {
		return
	}

	buff := &bytes.Buffer{}
	err = ByteWriter{Writer: buff}.
		Write([]*yaml.RNode{node2, node3, node1})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `c: d # second
---
e: f
g:
  h:
  - i # has a list
  - j
---
a: b #first
`, buff.String())
}

// TestByteWriter_Write_withAnnotationsKeepAnnotations tests:
// - Resource Config is sorted by annotations if present
// - IndexAnnotations are retained
func TestByteWriter_Write_withAnnotationsKeepAnnotations(t *testing.T) {
	node1, err := yaml.Parse(`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		return
	}
	node2, err := yaml.Parse(`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		return
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
		return
	}

	buff := &bytes.Buffer{}
	err = ByteWriter{Sort: true, Writer: buff, KeepReaderAnnotations: true}.
		Write([]*yaml.RNode{node2, node3, node1})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/b_test.yaml"
`, buff.String())
}

// TestByteWriter_Write_withAnnotations tests:
// - Resource Config is sorted by annotations if present
// - IndexAnnotations are pruned
func TestByteWriter_Write_withAnnotations(t *testing.T) {
	node1, err := yaml.Parse(`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		return
	}
	node2, err := yaml.Parse(`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		return
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
		return
	}

	buff := &bytes.Buffer{}
	err = ByteWriter{Sort: true, Writer: buff}.
		Write([]*yaml.RNode{node2, node3, node1})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a: b #first
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/b_test.yaml"
`, buff.String())
}

// TestByteWriter_Write_partialValues tests:
// - Resource Config is sorted when annotations are present on some but not all ResourceNodes
func TestByteWriter_Write_partialAnnotations(t *testing.T) {
	node1, err := yaml.Parse(`a: b #first
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		return
	}
	node2, err := yaml.Parse(`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
`)
	if !assert.NoError(t, err) {
		return
	}
	node3, err := yaml.Parse(`e: f
g:
  h:
  - i # has a list
  - j
`)
	if !assert.NoError(t, err) {
		return
	}

	buff := &bytes.Buffer{}
	rw := ByteWriter{Sort: true, Writer: buff}
	err = rw.Write([]*yaml.RNode{node2, node3, node1})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `e: f
g:
  h:
  - i # has a list
  - j
---
a: b #first
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
`, buff.String())
}
