// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/internal/wrappy"
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
	testDeploymentSpec := map[string]interface{}{
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"hostAliases": []interface{}{
					map[string]interface{}{
						"hostnames": []interface{}{
							"a.example.com",
						},
						"ip": "8.8.8.8",
					},
				},
			},
		},
	}
	testDeploymentA := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "deployment-a",
		},
		"spec": testDeploymentSpec,
	}
	testDeploymentB := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "deployment-b",
		},
		"spec": testDeploymentSpec,
	}
	testDeploymentList :=
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "DeploymentList",
			"items": []interface{}{
				testDeploymentA,
				testDeploymentB,
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
		"listWithAnchors": {
			input: []byte(`
apiVersion: v1
kind: DeploymentList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-a
  spec: &hostAliases
    template:
      spec:
        hostAliases:
        - hostnames:
          - a.example.com
          ip: 8.8.8.8
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-b
  spec:
    <<: *hostAliases
`),
			exp: expected{
				out: []map[string]interface{}{testDeploymentList},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rs, err := factory.SliceFromBytes(tc.input)
			if err != nil {
				assert.True(t, tc.exp.isErr)
				return
			}
			assert.False(t, tc.exp.isErr)
			assert.Equal(t, len(tc.exp.out), len(rs))
			for i := range rs {
				assert.Equal(
					t, fmt.Sprintf("%v", tc.exp.out[i]), fmt.Sprintf("%v", rs[i].Map()))
				if n != "listWithAnchors" {
					// https://github.com/kubernetes-sigs/kustomize/issues/3271
					if !reflect.DeepEqual(tc.exp.out[i], rs[i].Map()) {
						t.Fatalf("%s:\nexpected: %v\n  actual: %v",
							n, tc.exp.out[i], rs[i].Map())
					}
				}
			}
		})
	}
}
