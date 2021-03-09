// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/loader"
	. "sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestRNodesFromBytes(t *testing.T) {
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
			result, err := factory.RNodesFromBytes([]byte(tc.input))
			if err != nil {
				t.Fatalf("%v: fails with err: %v", name, err)
			}
			if len(result) != len(tc.expected) {
				for i := range result {
					str, err := result[i].String()
					if err != nil {
						t.Fatalf("%v: result to YAML fails with err: %v", name, err)
					}
					t.Logf("--- %d:\n%s", i, str)
				}
				t.Fatalf(
					"%v: actual len %d != expected len %d",
					name, len(result), len(tc.expected))
			}
			for i := range tc.expected {
				str, err := result[i].String()
				if err != nil {
					t.Fatalf("%v: result to YAML fails with err: %v", name, err)
				}
				if str != tc.expected[i] {
					t.Fatalf(
						"%v: string mismatch in item %d\n"+
							"actual:\n-----\n%s\n-----\n"+
							"expected:\n-----\n%s\n-----\n",
						name, i, str, tc.expected[i])
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
		loader.RestrictionRootOnly, filesys.Separator, fSys, false)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		input       []types.PatchStrategicMerge
		expectedOut []*Resource
		expectedErr bool
	}{
		"happy": {
			input:       []types.PatchStrategicMerge{patchGood1, patchGood2},
			expectedOut: []*Resource{testDeployment, testConfigMap},
			expectedErr: false,
		},
		"badFileName": {
			input:       []types.PatchStrategicMerge{patchGood1, "doesNotExist"},
			expectedOut: []*Resource{},
			expectedErr: true,
		},
		"badData": {
			input:       []types.PatchStrategicMerge{patchGood1, patchBad},
			expectedOut: []*Resource{},
			expectedErr: true,
		},
		"listOfPatches": {
			input:       []types.PatchStrategicMerge{patchList},
			expectedOut: []*Resource{testDeployment, testConfigMap},
			expectedErr: false,
		},
		"listWithAnchorReference": {
			input:       []types.PatchStrategicMerge{patchList2},
			expectedOut: []*Resource{testDeploymentA, testDeploymentB},
			// The error using kyaml is:
			//   json: unsupported type: map[interface {}]interface {}
			// maybe arising from too many conversions between
			// yaml, json, Resource, RNode, etc.
			// These conversions go away after closing #3506
			// TODO(#3271) This shouldn't have an error, but does when kyaml is used.
			expectedErr: true,
		},
		"listWithNoEntries": {
			input:       []types.PatchStrategicMerge{patchList3},
			expectedOut: []*Resource{},
			expectedErr: false,
		},
		"listWithNoItems": {
			input:       []types.PatchStrategicMerge{patchList4},
			expectedOut: []*Resource{},
			expectedErr: false,
		},
	}
	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			rs, err := factory.SliceFromPatches(ldr, test.input)
			if err != nil {
				assert.True(t, test.expectedErr,
					fmt.Sprintf("in test %s, got unexpected error: %v", n, err))
				return
			}
			assert.False(t, test.expectedErr, "expected no error in "+n)
			assert.Equal(t, len(test.expectedOut), len(rs))
			for i := range rs {
				expYaml, err := test.expectedOut[i].AsYAML()
				assert.NoError(t, err)
				actYaml, err := rs[i].AsYAML()
				assert.NoError(t, err)
				assert.Equal(t, expYaml, actYaml)
			}
		})
	}
}

func TestHash(t *testing.T) {
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
	k, err := factory.SliceFromBytes([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	result, err := k[0].Hash(factory.Hasher())
	if err != nil {
		t.Fatal(err)
	}
	if result != expect {
		t.Fatalf("expect %s but got %s", expect, result)
	}
}

func TestMoreRNodesFromBytes(t *testing.T) {
	type expected struct {
		out   []string
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
			exp:   expected{},
		},
		"goodJson": {
			input: []byte(`
{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}
`),
			exp: expected{
				out: []string{
					`{"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": "winnie"}}`,
				},
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
				out: []string{`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`},
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
				out: []string{`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`},
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
				out: []string{},
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
				out: []string{`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`,
				},
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
				out: []string{
					`{"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": "winnie"}}`,
					`{"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": "winnie"}}`,
				},
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
  spec: &foo
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
    *foo
`),
			exp: expected{
				out: []string{
					`{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": {"name": "deployment-a"}, ` +
						`"spec": {"template": {"spec": {"hostAliases": [{"hostnames": ["a.example.com"], "ip": "8.8.8.8"}]}}}}`,
					`{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": {"name": "deployment-b"}, ` +
						`"spec": {"template": {"spec": {"hostAliases": [{"hostnames": ["a.example.com"], "ip": "8.8.8.8"}]}}}}`},
			},
		},
		"simpleAnchor": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: wildcard
data:
  color: &color-used blue
  feeling: *color-used
`),
			exp: expected{
				out: []string{`
apiVersion: v1
kind: ConfigMap
metadata:
  name: wildcard
data:
  color: blue
  feeling: blue
`},
			},
		},
	}
	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rs, err := factory.RNodesFromBytes(tc.input)
			if err != nil {
				assert.True(t, tc.exp.isErr)
				return
			}
			assert.False(t, tc.exp.isErr)
			assert.Equal(t, len(tc.exp.out), len(rs))
			for i := range rs {
				actual, err := rs[i].String()
				assert.NoError(t, err)
				assert.Equal(
					t, strings.TrimSpace(tc.exp.out[i]), strings.TrimSpace(actual))
			}
		})
	}
}

func TestDropLocalNodes(t *testing.T) {
	testCases := map[string]struct {
		input    []byte
		expected []byte
	}{
		"localConfigUnset": {
			input: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			expected: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
		},
		"localConfigSet": {
			input: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie-skip
  annotations:
     # this annotation causes the Resource to be ignored by kustomize
     config.kubernetes.io/local-config: ""
`),
			expected: nil,
		},
		"localConfigSetToTrue": {
			input: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie-skip
  annotations:
	 config.kubernetes.io/local-config: "true"
		`),
			expected: nil,
		},
		"localConfigSetToFalse": {
			input: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
  annotations:
    config.kubernetes.io/local-config: "false"
`),
			expected: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/local-config: "false"
  name: winnie
`),
		},
		"localConfigMultiInput": {
			input: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie-skip
  annotations:
    config.kubernetes.io/local-config: "true"
`),
			expected: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
		},
	}
	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			nin, _ := kio.FromBytes(tc.input)
			res, err := factory.DropLocalNodes(nin)
			assert.NoError(t, err)
			if tc.expected == nil {
				assert.Equal(t, 0, len(res))
			} else {
				actual, _ := res[0].AsYAML()
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
