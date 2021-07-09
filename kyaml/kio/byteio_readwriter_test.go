// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestByteReadWriter(t *testing.T) {
	type testCase struct {
		name           string
		err            string
		input          string
		expectedOutput string
		instance       kio.ByteReadWriter
	}

	testCases := []testCase{
		{
			name: "round_trip",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`,
		},

		{
			name: "function_config",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
functionConfig:
  a: b # something
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
functionConfig:
  a: b # something
`,
		},

		{
			name: "results",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
results:
  a: b # something
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
results:
  a: b # something
`,
		},

		{
			name: "drop_invalid_resource_list_field",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
foo:
  a: b # something
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
`,
		},

		{
			name: "list",
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
			expectedOutput: `
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
		},

		{
			name: "multiple_documents",
			input: `
kind: Deployment
spec:
  replicas: 1
---
kind: Service
spec:
  selectors:
    foo: bar
`,
			expectedOutput: `
kind: Deployment
spec:
  replicas: 1
---
kind: Service
spec:
  selectors:
    foo: bar
`,
		},

		{
			name: "keep_annotations",
			input: `
kind: Deployment
spec:
  replicas: 1
---
kind: Service
spec:
  selectors:
    foo: bar
`,
			expectedOutput: `
kind: Deployment
spec:
  replicas: 1
metadata:
  annotations:
    config.kubernetes.io/index: '0'
---
kind: Service
spec:
  selectors:
    foo: bar
metadata:
  annotations:
    config.kubernetes.io/index: '1'
`,
			instance: kio.ByteReadWriter{KeepReaderAnnotations: true},
		},

		{
			name: "manual_override_wrap",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
functionConfig:
  a: b # something
`,
			expectedOutput: `
kind: Deployment
spec:
  replicas: 1
---
kind: Service
spec:
  selectors:
    foo: bar
`,
			instance: kio.ByteReadWriter{NoWrap: true},
		},

		{
			name: "manual_override_function_config",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
functionConfig:
  a: b # something
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  spec:
    replicas: 1
- kind: Service
  spec:
    selectors:
      foo: bar
functionConfig:
  c: d
`,
			instance: kio.ByteReadWriter{FunctionConfig: yaml.MustParse(`c: d`)},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			var in, out bytes.Buffer
			in.WriteString(tc.input)
			w := tc.instance
			w.Writer = &out
			w.Reader = &in

			nodes, err := w.Read()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = w.Write(nodes)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if tc.err != "" {
				if !assert.EqualError(t, err, tc.err) {
					t.FailNow()
				}
				return
			}

			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String())) {
				t.FailNow()
			}
		})
	}
}

func TestByteReadWriter_RetainSeqIndent(t *testing.T) {
	type testCase struct {
		name           string
		err            string
		input          string
		expectedOutput string
		instance       kio.ByteReadWriter
	}

	testCases := []testCase{
		{
			name: "round_trip with 2 space seq indent",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
---
apiVersion: v1
kind: Service
spec:
  - foo
  - bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
---
apiVersion: v1
kind: Service
spec:
  - foo
  - bar
`,
		},
		{
			name: "round_trip with 0 space seq indent",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
---
apiVersion: v1
kind: Service
spec:
- foo
- bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
---
apiVersion: v1
kind: Service
spec:
- foo
- bar
`,
		},
		{
			name: "round_trip with different indentations",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
---
apiVersion: v1
kind: Service
spec:
- foo
- bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
---
apiVersion: v1
kind: Service
spec:
- foo
- bar
`,
		},
		{
			name: "round_trip with mixed indentations in same resource, wide wins",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
env:
- foo
- bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
env:
  - foo
  - bar
`,
		},
		{
			name: "round_trip with mixed indentations in same resource, compact wins",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
env:
  - foo
  - bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
env:
- foo
- bar
`,
		},
		{
			name: "round_trip with mixed indentations in same resource, compact in case of a tie",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
env:
  - foo
  - bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
env:
- foo
- bar
`,
		},
		{
			name: "unwrap ResourceList with annotations",
			input: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
  - kind: Deployment
    metadata:
      annotations:
        internal.config.kubernetes.io/seqindent: "compact"
    spec:
      - foo
      - bar
  - kind: Service
    metadata:
      annotations:
        internal.config.kubernetes.io/seqindent: "wide"
    spec:
      - foo
      - bar
`,
			expectedOutput: `
kind: Deployment
spec:
- foo
- bar
---
kind: Service
spec:
  - foo
  - bar
`,
		},
		{
			name: "round_trip with mixed indentations in same resource, wide wins as it is first",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
env:
- foo
- bar
- baz
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
env:
  - foo
  - bar
  - baz
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			var in, out bytes.Buffer
			in.WriteString(tc.input)
			w := tc.instance
			w.Writer = &out
			w.Reader = &in
			w.PreserveSeqIndent = true

			nodes, err := w.Read()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w.WrappingKind = ""
			err = w.Write(nodes)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if tc.err != "" {
				if !assert.EqualError(t, err, tc.err) {
					t.FailNow()
				}
				return
			}

			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String())) {
				t.FailNow()
			}
		})
	}
}
