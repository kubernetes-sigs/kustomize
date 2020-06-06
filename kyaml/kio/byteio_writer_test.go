// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestByteWriter(t *testing.T) {
	type testCase struct {
		name           string
		err            string
		items          []string
		functionConfig string
		results        string
		expectedOutput string
		instance       ByteWriter
	}

	testCases := []testCase{
		//
		//
		//
		{
			name: "wrap_resource_list",
			instance: ByteWriter{
				Sort:               true,
				WrappingKind:       ResourceListKind,
				WrappingAPIVersion: ResourceListAPIVersion,
			},
			items: []string{
				`a: b #first`,
				`c: d # second`,
			},
			functionConfig: `
e: f
g:
  h:
  - i # has a list
  - j`,
			expectedOutput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- a: b #first
- c: d # second
functionConfig:
  e: f
  g:
    h:
    - i # has a list
    - j
`,
		},

		//
		//
		//
		{
			name: "multiple_items",
			items: []string{
				`c: d # second`,
				`e: f
g:
  h:
  # has a list
  - i : [i1, i2] # line comment
  # has a list 2
  - j : j1
`,
				`a: b #first`,
			},
			expectedOutput: `
c: d # second
---
e: f
g:
  h:
  # has a list
  - i: [i1, i2] # line comment
  # has a list 2
  - j: j1
---
a: b #first
`,
		},

		//
		// Test Case
		//
		{
			name:     "sort_keep_annotation",
			instance: ByteWriter{Sort: true, KeepReaderAnnotations: true},
			items: []string{
				`a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/a_test.yaml"
`,
				`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/b_test.yaml"
`,
				`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
`,
			},

			expectedOutput: `a: b #first
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    config.kubernetes.io/index: 0
    config.kubernetes.io/path: "a/b/b_test.yaml"
`,
		},

		//
		// Test Case
		//
		{
			name:     "sort_partial_annotations",
			instance: ByteWriter{Sort: true},
			items: []string{
				`a: b #first
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
`,
				`c: d # second
metadata:
  annotations:
    config.kubernetes.io/index: 1
    config.kubernetes.io/path: "a/b/a_test.yaml"
`,
				`e: f
g:
  h:
  - i # has a list
  - j
`,
			},

			expectedOutput: `e: f
g:
  h:
  - i # has a list
  - j
---
a: b #first
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    config.kubernetes.io/path: "a/b/a_test.yaml"
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			actual := &bytes.Buffer{}
			w := tc.instance
			w.Writer = actual

			if tc.functionConfig != "" {
				w.FunctionConfig = yaml.MustParse(tc.functionConfig)
			}

			if tc.results != "" {
				w.Results = yaml.MustParse(tc.results)
			}

			var items []*yaml.RNode
			for i := range tc.items {
				items = append(items, yaml.MustParse(tc.items[i]))
			}
			err := w.Write(items)

			if tc.err != "" {
				if !assert.EqualError(t, err, tc.err) {
					t.FailNow()
				}
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(actual.String())) {
				t.FailNow()
			}
		})
	}
}
