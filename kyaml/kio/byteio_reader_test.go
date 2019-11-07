// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
)

// getByteReaderTestInput returns test input
func getByteReaderTestInput(t *testing.T) *bytes.Buffer {
	b := &bytes.Buffer{}
	_, err := b.WriteString(`
---
a: b # first resource
c: d
---
# second resource
e: f
g:
- h
---
---
i: j
`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "")
	}
	return b
}

func TestByteReader_Read_wrappedResourceßßList(t *testing.T) {
	r := &ByteReader{Reader: bytes.NewBufferString(`apiVersion: kyaml.kustomize.dev/v1alpha1
kind: ResourceList
functionConfig:
  foo: bar
  elems:
  - a
  - b
  - c
items:
-  kind: Deployment
   spec:
     replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`)}
	nodes, err := r.Read()
	if !assert.NoError(t, err) {
		return
	}

	// verify the contents
	if !assert.Len(t, nodes, 2) {
		return
	}
	expected := []string{
		`kind: Deployment
spec:
  replicas: 1
`,
		`kind: Service
spec:
  selectors:
    foo: bar
`,
	}
	for i := range nodes {
		if !assert.Equal(t, expected[i], nodes[i].MustString()) {
			return
		}
	}

	// verify the function config
	assert.Equal(t, `foo: bar
elems:
- a
- b
- c
`, r.FunctionConfig.MustString())

	assert.Equal(t, ResourceListKind, r.WrappingKind)
	assert.Equal(t, ResourceListApiVersion, r.WrappingApiVersion)

}

func TestByteReader_Read_wrappedList(t *testing.T) {
	r := &ByteReader{Reader: bytes.NewBufferString(`apiVersion: v1
kind: List
items:
-  kind: Deployment
   spec:
     replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`)}
	nodes, err := r.Read()
	if !assert.NoError(t, err) {
		return
	}

	// verify the contents
	if !assert.Len(t, nodes, 2) {
		return
	}
	expected := []string{
		`kind: Deployment
spec:
  replicas: 1
`,
		`kind: Service
spec:
  selectors:
    foo: bar
`,
	}
	for i := range nodes {
		if !assert.Equal(t, expected[i], nodes[i].MustString()) {
			return
		}
	}

	// verify the function config
	assert.Nil(t, r.FunctionConfig)
	assert.Equal(t, "List", r.WrappingKind)
	assert.Equal(t, "v1", r.WrappingApiVersion)
}

// TestByteReader_Read tests the default Read behavior
// - Resources are read into a slice
// - ReaderAnnotations are set on the ResourceNodes
func TestByteReader_Read(t *testing.T) {
	nodes, err := (&ByteReader{Reader: getByteReaderTestInput(t)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Len(t, nodes, 3) {
		return
	}
	expected := []string{
		`a: b # first resource
c: d
metadata:
  annotations:
    config.kubernetes.io/index: 0
`,
		`# second resource
e: f
g:
- h
metadata:
  annotations:
    config.kubernetes.io/index: 1
`,
		`i: j
metadata:
  annotations:
    config.kubernetes.io/index: 2
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

// TestByteReader_Read_omitReaderAnnotations tests
// - Resources are read into a slice
// - ReaderAnnotations are not set on the ResourceNodes
func TestByteReader_Read_omitReaderAnnotations(t *testing.T) {
	nodes, err := (&ByteReader{
		Reader:                getByteReaderTestInput(t),
		OmitReaderAnnotations: true}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// should have parsed 3 resources
	if !assert.Len(t, nodes, 3) {
		return
	}
	expected := []string{
		"a: b # first resource\nc: d\n",
		"# second resource\ne: f\ng:\n- h\n",
		"i: j\n",
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

// TestByteReader_Read_omitReaderAnnotations tests
// - Resources are read into a slice
// - ReaderAnnotations are NOT set on the ResourceNodes
// - Additional annotations ARE set on the ResourceNodes
func TestByteReader_Read_setAnnotationsOmitReaderAnnotations(t *testing.T) {
	nodes, err := (&ByteReader{
		Reader:                getByteReaderTestInput(t),
		SetAnnotations:        map[string]string{"foo": "bar"},
		OmitReaderAnnotations: true,
	}).Read()
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Len(t, nodes, 3) {
		return
	}
	expected := []string{
		`a: b # first resource
c: d
metadata:
  annotations:
    foo: bar
`,
		`# second resource
e: f
g:
- h
metadata:
  annotations:
    foo: bar
`,
		`i: j
metadata:
  annotations:
    foo: bar
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

// TestByteReader_Read_omitReaderAnnotations tests
// - Resources are read into a slice
// - ReaderAnnotations ARE set on the ResourceNodes
// - Additional annotations ARE set on the ResourceNodes
func TestByteReader_Read_setAnnotations(t *testing.T) {
	nodes, err := (&ByteReader{
		Reader:         getByteReaderTestInput(t),
		SetAnnotations: map[string]string{"foo": "bar"},
	}).Read()
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Len(t, nodes, 3) {
		return
	}
	expected := []string{
		`a: b # first resource
c: d
metadata:
  annotations:
    config.kubernetes.io/index: 0
    foo: bar
`,
		`# second resource
e: f
g:
- h
metadata:
  annotations:
    config.kubernetes.io/index: 1
    foo: bar
`,
		`i: j
metadata:
  annotations:
    config.kubernetes.io/index: 2
    foo: bar
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
