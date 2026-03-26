// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	. "sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestStripCommentsFilter(t *testing.T) {
	input := `
# this is a head comment
apiVersion: apps/v1 # this is a line comment
kind: Deployment
metadata:
  name: foo # name comment
  # annotation comment
  namespace: bar
`
	node, err := yaml.Parse(input)
	assert.NoError(t, err)

	result, err := StripCommentsFilter{}.Filter([]*yaml.RNode{node})
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	out, err := result[0].String()
	assert.NoError(t, err)
	assert.NotContains(t, out, "head comment")
	assert.NotContains(t, out, "line comment")
	assert.NotContains(t, out, "name comment")
	assert.NotContains(t, out, "annotation comment")
}

func TestStripCommentsFilter_Empty(t *testing.T) {
	result, err := StripCommentsFilter{}.Filter(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestStripCommentsFilter_MultipleResources(t *testing.T) {
	input := `
# comment 1
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1 # inline
---
# comment 2
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`
	in := bytes.NewBufferString(input)
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: in}},
		Filters: []Filter{StripCommentsFilter{}},
		Outputs: []Writer{ByteWriter{Writer: out}},
	}.Execute()
	assert.NoError(t, err)
	assert.NotContains(t, out.String(), "comment 1")
	assert.NotContains(t, out.String(), "comment 2")
	assert.NotContains(t, out.String(), "inline")
	assert.Contains(t, out.String(), "name: cm1")
	assert.Contains(t, out.String(), "name: cm2")
}
