// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.

package resmap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/resid"
	. "sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
)

func setupRMForPatchTargets(t *testing.T) ResMap {
	result, err := rmF.NewResMapFromBytes([]byte(`
apiVersion: group1/v1
kind: Kind1
metadata:
  name: name1
  namespace: ns1
  labels:
    app: name1
  annotations:
    foo: bar
---
apiVersion: group1/v1
kind: Kind1
metadata:
  name: name2
  namespace: default
  labels:
    app: name2
  annotations:
    foo: bar
---
apiVersion: group1/v1
kind: Kind2
metadata:
  name: name3
  labels:
    app: name3
  annotations:
    bar: baz
---
apiVersion: group1/v1
kind: Kind2
metadata:
  name: x-name1
  namespace: x-default
`))
	assert.NoError(t, err)
	return result
}

func TestFindPatchTargets(t *testing.T) {
	rm := setupRMForPatchTargets(t)
	testcases := map[string]struct {
		target types.Selector
		count  int
	}{
		"select_01": {
			target: types.Selector{
				Name: "name.*",
			},
			count: 3,
		},
		"select_02": {
			target: types.Selector{
				Name:               "name.*",
				AnnotationSelector: "foo=bar",
			},
			count: 2,
		},
		"select_03": {
			target: types.Selector{
				LabelSelector: "app=name1",
			},
			count: 1,
		},
		"select_04": {
			target: types.Selector{
				Gvk: resid.Gvk{
					Kind: "Kind1",
				},
				Name: "name.*",
			},
			count: 2,
		},
		"select_05": {
			target: types.Selector{
				Name: "NotMatched",
			},
			count: 0,
		},
		"select_06": {
			target: types.Selector{
				Name: "",
			},
			count: 4,
		},
		"select_07": {
			target: types.Selector{
				Namespace: "default",
			},
			count: 2,
		},
		"select_08": {
			target: types.Selector{
				Namespace: "",
			},
			count: 4,
		},
		"select_09": {
			target: types.Selector{
				Namespace: "default",
				Name:      "name.*",
				Gvk: resid.Gvk{
					Kind: "Kind1",
				},
			},
			count: 1,
		},
		"select_10": {
			target: types.Selector{
				Name: "^name.*",
			},
			count: 3,
		},
		"select_11": {
			target: types.Selector{
				Name: "name.*$",
			},
			count: 3,
		},
		"select_12": {
			target: types.Selector{
				Name: "^name.*$",
			},
			count: 3,
		},
		"select_13": {
			target: types.Selector{
				Namespace: "^def.*",
			},
			count: 2,
		},
		"select_14": {
			target: types.Selector{
				Namespace: "def.*$",
			},
			count: 2,
		},
		"select_15": {
			target: types.Selector{
				Namespace: "^def.*$",
			},
			count: 2,
		},
		"select_16": {
			target: types.Selector{
				Namespace: "default",
			},
			count: 2,
		},
		"select_17": {
			target: types.Selector{
				Namespace: "NotMatched",
			},
			count: 0,
		},
		"select_18": {
			target: types.Selector{
				Namespace: "ns1",
			},
			count: 1,
		},
	}
	for n, testcase := range testcases {
		actual, err := rm.Select(testcase.target)
		assert.NoError(t, err)
		assert.Equalf(
			t, testcase.count, len(actual), "test=%s target=%v", n, testcase.target)
	}
}
