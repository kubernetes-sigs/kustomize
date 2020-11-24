// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
)

func TestByteReader(t *testing.T) {
	type testCase struct {
		name                   string
		input                  string
		err                    string
		expectedItems          []string
		expectedFunctionConfig string
		expectedResults        string
		wrappingAPIVersion     string
		wrappingAPIKind        string
		instance               ByteReader
	}

	testCases := []testCase{
		//
		//
		//
		{
			name: "wrapped_resource_list",
			input: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
-  kind: Deployment
   spec:
     replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`,
			expectedItems: []string{
				`kind: Deployment
spec:
  replicas: 1
`,
				`kind: Service
spec:
  selectors:
    foo: bar
`,
			},
			wrappingAPIVersion: ResourceListAPIVersion,
			wrappingAPIKind:    ResourceListKind,
		},

		//
		//
		//
		{
			name: "wrapped_resource_list_function_config",
			input: `apiVersion: config.kubernetes.io/v1alpha1
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
`,
			expectedItems: []string{
				`kind: Deployment
spec:
  replicas: 1
`,
				`kind: Service
spec:
  selectors:
    foo: bar
`,
			},
			expectedFunctionConfig: `foo: bar
elems:
- a
- b
- c`,
			wrappingAPIVersion: ResourceListAPIVersion,
			wrappingAPIKind:    ResourceListKind,
		},

		//
		//
		//
		{
			name: "wrapped_list",
			input: `
apiVersion: v1
kind: List
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`,
			expectedItems: []string{
				`
kind: Deployment
spec:
  replicas: 1
`,
				`
kind: Service
spec:
  selectors:
    foo: bar
`,
			},
			wrappingAPIKind:    "List",
			wrappingAPIVersion: "v1",
		},

		//
		//
		//
		{
			name: "unwrapped_items",
			input: `
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
`,
			expectedItems: []string{
				`a: b # first resource
c: d
metadata:
  annotations:
    config.kubernetes.io/index: '0'
`,
				`# second resource
e: f
g:
- h
metadata:
  annotations:
    config.kubernetes.io/index: '1'
`,
				`i: j
metadata:
  annotations:
    config.kubernetes.io/index: '2'
`,
			},
		},

		//
		//
		//
		{
			name: "omit_annotations",
			input: `
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
`,
			expectedItems: []string{
				`
a: b # first resource
c: d
`,
				`
# second resource
e: f
g:
- h
`,
				`
i: j
`,
			},
			instance: ByteReader{OmitReaderAnnotations: true},
		},

		//
		//
		//
		{
			name: "no_omit_annotations",
			input: `
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
`,
			expectedItems: []string{
				`
a: b # first resource
c: d
metadata:
  annotations:
    config.kubernetes.io/index: '0'
`,
				`
# second resource
e: f
g:
- h
metadata:
  annotations:
    config.kubernetes.io/index: '1'
`,
				`
i: j
metadata:
  annotations:
    config.kubernetes.io/index: '2'
`,
			},
			instance: ByteReader{},
		},

		//
		//
		//
		{
			name: "set_annotation",
			input: `
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
`,
			expectedItems: []string{
				`a: b # first resource
c: d
metadata:
  annotations:
    foo: 'bar'
`,
				`# second resource
e: f
g:
- h
metadata:
  annotations:
    foo: 'bar'
`,
				`i: j
metadata:
  annotations:
    foo: 'bar'
`,
			},
			instance: ByteReader{
				OmitReaderAnnotations: true,
				SetAnnotations:        map[string]string{"foo": "bar"}},
		},

		//
		//
		//
		{
			name:  "windows_line_ending",
			input: "\r\n---\r\na: b # first resource\r\nc: d\r\n---\r\n# second resource\r\ne: f\r\ng:\r\n- h\r\n---\r\n\r\n---\r\n i: j",
			expectedItems: []string{
				`a: b # first resource
c: d
metadata:
  annotations:
    foo: 'bar'
`,
				`# second resource
e: f
g:
- h
metadata:
  annotations:
    foo: 'bar'
`,
				`i: j
metadata:
  annotations:
    foo: 'bar'
`,
			},
			instance: ByteReader{
				OmitReaderAnnotations: true,
				SetAnnotations:        map[string]string{"foo": "bar"}},
		},

		//
		//
		//
		{
			name: "json",
			input: `
{
  "a": "b",
  "c": [1, 2]
}
`,
			expectedItems: []string{
				`
{"a": "b", "c": [1, 2], metadata: {annotations: {config.kubernetes.io/index: '0'}}}
`,
			},
			instance: ByteReader{},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			r := tc.instance
			r.Reader = bytes.NewBufferString(tc.input)
			nodes, err := r.Read()
			if tc.err != "" {
				if !assert.EqualError(t, err, tc.err) {
					t.FailNow()
				}
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// verify the contents
			if !assert.Len(t, nodes, len(tc.expectedItems)) {
				t.FailNow()
			}
			for i := range nodes {
				actual, err := nodes[i].String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tc.expectedItems[i]),
					strings.TrimSpace(actual)) {
					t.FailNow()
				}
			}

			// verify the function config
			if tc.expectedFunctionConfig != "" {
				actual, err := r.FunctionConfig.String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tc.expectedFunctionConfig),
					strings.TrimSpace(actual)) {
					t.FailNow()
				}
			} else if !assert.Nil(t, r.FunctionConfig) {
				t.FailNow()
			}

			if tc.expectedResults != "" {
				actual, err := r.Results.String()
				actual = strings.TrimSpace(actual)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				tc.expectedResults = strings.TrimSpace(tc.expectedResults)
				if !assert.Equal(t, tc.expectedResults, actual) {
					t.FailNow()
				}
			} else if !assert.Nil(t, r.Results) {
				t.FailNow()
			}

			if !assert.Equal(t, tc.wrappingAPIKind, r.WrappingKind) {
				t.FailNow()
			}
			if !assert.Equal(t, tc.wrappingAPIVersion, r.WrappingAPIVersion) {
				t.FailNow()
			}
		})
	}
}

func TestFromBytes(t *testing.T) {
	type expected struct {
		isErr bool
		sOut  []string
	}

	testCases := map[string]struct {
		input []byte
		exp   expected
	}{
		"garbage": {
			input: []byte("garbageIn: garbageOut"),
			exp: expected{
				sOut: []string{"garbageIn: garbageOut"},
			},
		},
		"noBytes": {
			input: []byte{},
			exp: expected{
				sOut: []string{},
			},
		},
		"goodJson": {
			input: []byte(`
{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}
`),
			exp: expected{
				sOut: []string{`
{"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": "winnie"}}
`},
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
				sOut: []string{`
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
				sOut: []string{`
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
				sOut: []string{`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie-skip
  annotations:
    # this annotation causes the Resource to be ignored by kustomize
    config.kubernetes.io/local-config: ""
`,
					`
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
WOOOOOOOOOOOOOOOOOOOOOOOOT:     woot
`),
			exp: expected{
				sOut: []string{`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`,
					`
WOOOOOOOOOOOOOOOOOOOOOOOOT: woot
`},
			},
		},
		"emptyObjects": {
			input: []byte(`
---
#a comment

---

`),
			exp: expected{
				sOut: []string{},
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
				sOut: []string{`
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    foo: bar
`},
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
				sOut: []string{`
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
				sOut: []string{`
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
`},
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
				sOut: []string{`
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
    !!merge <<: *hostAliases
`},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rNodes, err := FromBytes(tc.input)
			if err != nil {
				assert.True(t, tc.exp.isErr)
				return
			}
			assert.False(t, tc.exp.isErr)
			assert.Equal(t, len(tc.exp.sOut), len(rNodes))
			for i, n := range rNodes {
				json, err := n.String()
				assert.NoError(t, err)
				assert.Equal(
					t, strings.TrimSpace(tc.exp.sOut[i]),
					strings.TrimSpace(json), n)
			}
		})
	}
}
