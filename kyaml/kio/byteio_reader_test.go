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
