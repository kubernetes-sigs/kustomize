// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/loader"
	. "sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
)

func TestSliceFromBytes(t *testing.T) {
	type testCase struct {
		input    string
		expected []string
	}
	testCases := map[string]testCase{
		"empty1": {
			input:    "",
			expected: []string{},
		},
		"empty2": {
			input: `
---
---
`,
			expected: []string{},
		},
		"deployment1": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
---
`,
			expected: []string{
				`apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
`,
			},
		},
		"deployment2": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: baseprefix-mungebot
spec:
  replicas: 1
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      annotations:
        baseAnno: This is a base annotation
      labels:
        app: mungebot
        foo: bar
    spec:
      containers:
      - env:
        - name: foo
          value: bar
        image: nginx
        name: nginx
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: baseprefix-mungebot-service
spec:
  ports:
  - port: 7002
  selector:
    app: mungebot
    foo: bar
`,
			expected: []string{
				`apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: baseprefix-mungebot
spec:
  replicas: 1
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      annotations:
        baseAnno: This is a base annotation
      labels:
        app: mungebot
        foo: bar
    spec:
      containers:
      - env:
        - name: foo
          value: bar
        image: nginx
        name: nginx
        ports:
        - containerPort: 80
`,
				`apiVersion: v1
kind: Service
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: baseprefix-mungebot-service
spec:
  ports:
  - port: 7002
  selector:
    app: mungebot
    foo: bar
`,
			},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			result, err := factory.SliceFromBytes([]byte(tc.input))
			if err != nil {
				t.Fatalf("%v: fails with err: %v", name, err)
			}
			if len(result) != len(tc.expected) {
				for i := range result {
					bytes, err := result[i].AsYAML()
					if err != nil {
						t.Fatalf("%v: result to YAML fails with err: %v", name, err)
					}
					tmp := string(bytes)
					t.Logf("--- %d:\n%s", i, tmp)
				}
				t.Fatalf(
					"%v: actual len %d != expected len %d",
					name, len(result), len(tc.expected))
			}
			for i := range tc.expected {
				bytes, err := result[i].AsYAML()
				if err != nil {
					t.Fatalf("%v: result to YAML fails with err: %v", name, err)
				}
				tmp := string(bytes)
				if tmp != tc.expected[i] {
					t.Fatalf(
						"%v: string mismatch in item %d\n"+
							"actual:\n-----\n%s\n-----\n"+
							"expected:\n-----\n%s\n-----\n",
						name, i, tmp, tc.expected[i])
				}
			}
		})
	}
}

func TestSliceFromPatches(t *testing.T) {
	patchGood1 := types.PatchStrategicMerge("patch1.yaml")
	patch1 := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
`
	patchGood2 := types.PatchStrategicMerge("patch2.yaml")
	patch2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
  namespace: hundred-acre-wood
---
# some comment
---
---
`
	patchBad := types.PatchStrategicMerge("patch3.yaml")
	patch3 := `
WOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOT: woot
`
	patchList := types.PatchStrategicMerge("patch4.yaml")
	patch4 := `
apiVersion: v1
kind: List
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: pooh
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
    namespace: hundred-acre-wood
`
	patchList2 := types.PatchStrategicMerge("patch5.yaml")
	patch5 := `
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
`
	patchList3 := types.PatchStrategicMerge("patch6.yaml")
	patch6 := `
apiVersion: v1
kind: List
items:
`
	patchList4 := types.PatchStrategicMerge("patch7.yaml")
	patch7 := `
apiVersion: v1
kind: List
`
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
	testDeploymentA := factory.FromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "deployment-a",
			},
			"spec": testDeploymentSpec,
		})
	testDeploymentB := factory.FromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "deployment-b",
			},
			"spec": testDeploymentSpec,
		})

	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(string(patchGood1), []byte(patch1))
	fSys.WriteFile(string(patchGood2), []byte(patch2))
	fSys.WriteFile(string(patchBad), []byte(patch3))
	fSys.WriteFile(string(patchList), []byte(patch4))
	fSys.WriteFile(string(patchList2), []byte(patch5))
	fSys.WriteFile(string(patchList3), []byte(patch6))
	fSys.WriteFile(string(patchList4), []byte(patch7))

	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		input       []types.PatchStrategicMerge
		expectedOut []*Resource
		expectedErr bool
	}{
		{
			name:        "happy",
			input:       []types.PatchStrategicMerge{patchGood1, patchGood2},
			expectedOut: []*Resource{testDeployment, testConfigMap},
			expectedErr: false,
		},
		{
			name:        "badFileName",
			input:       []types.PatchStrategicMerge{patchGood1, "doesNotExist"},
			expectedOut: []*Resource{},
			expectedErr: true,
		},
		{
			name:        "badData",
			input:       []types.PatchStrategicMerge{patchGood1, patchBad},
			expectedOut: []*Resource{},
			expectedErr: true,
		},
		{
			name:        "listOfPatches",
			input:       []types.PatchStrategicMerge{patchList},
			expectedOut: []*Resource{testDeployment, testConfigMap},
			expectedErr: false,
		},
		{
			name:        "listWithAnchorReference",
			input:       []types.PatchStrategicMerge{patchList2},
			expectedOut: []*Resource{testDeploymentA, testDeploymentB},
			expectedErr: false,
		},
		{
			name:        "listWithNoEntries",
			input:       []types.PatchStrategicMerge{patchList3},
			expectedOut: []*Resource{},
			expectedErr: false,
		},
		{
			name:        "listWithNo'items:'",
			input:       []types.PatchStrategicMerge{patchList4},
			expectedOut: []*Resource{},
			expectedErr: false,
		},
	}
	for _, test := range tests {
		rs, err := factory.SliceFromPatches(ldr, test.input)
		if test.expectedErr && err == nil {
			t.Fatalf("%v: should return error", test.name)
		}
		if !test.expectedErr && err != nil {
			t.Fatalf("%v: unexpected error: %s", test.name, err)
		}
		if len(rs) != len(test.expectedOut) {
			t.Fatalf("%s: length mismatch %d != %d",
				test.name, len(rs), len(test.expectedOut))
		}
		for i := range rs {
			if !reflect.DeepEqual(test.expectedOut[i], rs[i]) {
				t.Fatalf("%s: Got: %v\nexpected:%v",
					test.name, test.expectedOut[i], rs[i])
			}
		}
	}
}
