// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// TestMerge_Merge_order tests that the original order of elements
// retained after merge
func TestMerge_Merge_order(t *testing.T) {
	r1, err := yaml.Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-1
  namespace: bar-1
spec:
  template:
    spec: {}
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	r2, err := yaml.Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-2
  namespace: bar-2
spec:
  template:
    spec: {}
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var b bytes.Buffer
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.PackageBuffer{Nodes: []*yaml.RNode{r1, r2}}},
		Filters: []kio.Filter{filters.MatchFilter{}},
		Outputs: []kio.Writer{&kio.ByteWriter{Writer: &b}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := strings.TrimSpace(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-1
  namespace: bar-1
spec:
  template:
    spec: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-2
  namespace: bar-2
spec:
  template:
    spec: {}
`)

	if !assert.Equal(t, expected, strings.TrimSpace(b.String())) {
		t.FailNow()
	}

	b.Reset()
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.PackageBuffer{Nodes: []*yaml.RNode{r2, r1}}},
		Filters: []kio.Filter{filters.MatchFilter{}},
		Outputs: []kio.Writer{&kio.ByteWriter{Writer: &b}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected = strings.TrimSpace(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-2
  namespace: bar-2
spec:
  template:
    spec: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-1
  namespace: bar-1
spec:
  template:
    spec: {}
`)

	if !assert.Equal(t, expected, strings.TrimSpace(b.String())) {
		t.FailNow()
	}
}
