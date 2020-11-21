// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHasher(t *testing.T) {
	input := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: foo
data:
  one: ""
binaryData:
  two: ""
`
	expect := "698h7c7t9m"

	factory := &WNodeFactory{}
	k, err := factory.SliceFromBytes([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	hasher := factory.Hasher()
	result, err := hasher.Hash(k[0])
	if err != nil {
		t.Fatal(err)
	}
	if result != expect {
		t.Fatalf("expect %s but got %s", expect, result)
	}
}

func TestSliceFromBytes(t *testing.T) {
	factory := &WNodeFactory{}
	testConfigMap :=
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "winnie",
			},
		}
	testConfigMapList :=
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMapList",
			"items": []interface{}{
				testConfigMap,
				testConfigMap,
			},
		}

	type expected struct {
		out   []map[string]interface{}
		isErr bool
	}

	testCases := map[string]struct {
		input []byte
		exp   expected
	}{
		"garbage": {
			input: []byte("garbageIn: garbageOut"),
			exp: expected{
				isErr: true,
			},
		},
		"noBytes": {
			input: []byte{},
			exp: expected{
				out: []map[string]interface{}{},
			},
		},
		"goodJson": {
			input: []byte(`
{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap},
			},
		},
		"goodYaml1": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap},
			},
		},
		"goodYaml2": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap, testConfigMap},
			},
		},
		"localConfigYaml": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie-skip
  annotations:
    # this annotation causes the Resource to be ignored by kustomize
    config.kubernetes.io/local-config: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap},
			},
		},
		"garbageInOneOfTwoObjects": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
WOOOOOOOOOOOOOOOOOOOOOOOOT:  woot
`),
			exp: expected{
				isErr: true,
			},
		},
		"emptyObjects": {
			input: []byte(`
---
#a comment

---

`),
			exp: expected{
				out: []map[string]interface{}{},
			},
		},
		"Missing .metadata.name in object": {
			input: []byte(`
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    foo: bar
`),
			exp: expected{
				isErr: true,
			},
		},
		"nil value in list": {
			input: []byte(`
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: kube100-site
	labels:
	  app: web
testList:
- testA
-
`),
			exp: expected{
				isErr: true,
			},
		},
		"List": {
			input: []byte(`
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{
					testConfigMap,
					testConfigMap},
			},
		},
		"ConfigMapList": {
			input: []byte(`
apiVersion: v1
kind: ConfigMapList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMapList},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rs, err := factory.SliceFromBytes(tc.input)
			if tc.exp.isErr && err == nil {
				t.Fatalf("%v: should return error", n)
			}
			if !tc.exp.isErr && err != nil {
				t.Fatalf("%v: unexpected error: %s", n, err)
			}
			if len(tc.exp.out) != len(rs) {
				fmt.Printf("%s: \nexpected:%v\nactual: %v\n",
					n, tc.exp.out, rs)
				t.Fatalf("%s: length mismatch; expected %d, actual %d",
					n, len(tc.exp.out), len(rs))
			}
			for i := range rs {
				if !reflect.DeepEqual(tc.exp.out[i], rs[i].Map()) {
					t.Fatalf("%s: Got: %v\nexpected:%v",
						n, rs[i].Map(), tc.exp.out[i])
				}
			}
		})
	}
}
