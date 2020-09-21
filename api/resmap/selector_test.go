// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.

package resmap_test

import (
	"testing"

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
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	return result
}

func TestFindPatchTargets(t *testing.T) {
	rm := setupRMForPatchTargets(t)
	testcases := []struct {
		target types.Selector
		count  int
	}{
		{
			target: types.Selector{
				Name: "name.*",
			},
			count: 3,
		},
		{
			target: types.Selector{
				Name:               "name.*",
				AnnotationSelector: "foo=bar",
			},
			count: 2,
		},
		{
			target: types.Selector{
				LabelSelector: "app=name1",
			},
			count: 1,
		},
		{
			target: types.Selector{
				Gvk: resid.Gvk{
					Kind: "Kind1",
				},
				Name: "name.*",
			},
			count: 2,
		},
		{
			target: types.Selector{
				Name: "NotMatched",
			},
			count: 0,
		},
		{
			target: types.Selector{
				Name: "",
			},
			count: 4,
		},
		{
			target: types.Selector{
				Namespace: "default",
			},
			count: 2,
		},
		{
			target: types.Selector{
				Namespace: "",
			},
			count: 4,
		},
		{
			target: types.Selector{
				Namespace: "default",
				Name:      "name.*",
				Gvk: resid.Gvk{
					Kind: "Kind1",
				},
			},
			count: 1,
		},
		{
			target: types.Selector{
				Name: "^name.*",
			},
			count: 3,
		},
		{
			target: types.Selector{
				Name: "name.*$",
			},
			count: 3,
		},
		{
			target: types.Selector{
				Name: "^name.*$",
			},
			count: 3,
		},
		{
			target: types.Selector{
				Namespace: "^def.*",
			},
			count: 2,
		},
		{
			target: types.Selector{
				Namespace: "def.*$",
			},
			count: 2,
		},
		{
			target: types.Selector{
				Namespace: "^def.*$",
			},
			count: 2,
		},
		{
			target: types.Selector{
				Namespace: "default",
			},
			count: 2,
		},
		{
			target: types.Selector{
				Namespace: "NotMatched",
			},
			count: 0,
		},
		{
			target: types.Selector{
				Namespace: "ns1",
			},
			count: 1,
		},
	}
	for _, testcase := range testcases {
		actual, err := rm.Select(testcase.target)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if len(actual) != testcase.count {
			t.Errorf("expected %d objects, but got %d:\n%v", testcase.count, len(actual), actual)
		}
	}

}
