// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	. "sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestGrepFilter_Filter(t *testing.T) {
	in := `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
  name: bar
spec:
  replicas: 3
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`
	out := &bytes.Buffer{}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{GrepFilter{Path: []string{"metadata", "name"}, Value: "foo"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`, out.String()) {
		t.FailNow()
	}

	out = &bytes.Buffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{GrepFilter{Path: []string{"kind"}, Value: "Deployment"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
  name: bar
spec:
  replicas: 3
`, out.String()) {
		t.FailNow()
	}

	out = &bytes.Buffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{GrepFilter{Path: []string{"spec", "replicas"}, Value: "3"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
  name: bar
spec:
  replicas: 3
`, out.String()) {
		t.FailNow()
	}

	out = &bytes.Buffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{GrepFilter{Path: []string{"spec", "not-present"}, Value: "3"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, ``, out.String()) {
		t.FailNow()
	}
}

func TestGrepFilter_init(t *testing.T) {
	assert.Equal(t, GrepFilter{}, Filters["GrepFilter"]())
}

func TestGrepFilter_error(t *testing.T) {
	v, err := GrepFilter{Path: []string{"metadata", "name"},
		Value: "foo"}.Filter([]*yaml.RNode{{}})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Nil(t, v)
}
