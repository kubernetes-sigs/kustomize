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
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
			name: "results v1",
			input: `
apiVersion: config.kubernetes.io/v1
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
- field:
    proposedValue: "foo"
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v1
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
- field:
    proposedValue: "foo"
`,
		},

		{
			name: "results v1 proposedValues",
			err:  "ResourceList v1 does not support the array proposedValues, use the single proposedValue instead",
			input: `
apiVersion: config.kubernetes.io/v1
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
- field:
    proposedValues: ["foo"]
`,
		},

		{
			name: "results v2",
			input: `
apiVersion: config.kubernetes.io/v2
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
- field:
    proposedValues: ["foo", "bar"]
`,
			expectedOutput: `
apiVersion: config.kubernetes.io/v2
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
- field:
    proposedValues: ["foo", "bar"]
`,
		},

		{
			name: "results v2 proposedValue not supported",
			err:  "ResourceList v2alpha1 no longer supports a single proposedValue, use the proposedValues array instead",
			input: `
apiVersion: config.kubernetes.io/v2
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
- field:
    proposedValue: "foo"
`,
		},

		{
			name: "drop_invalid_resource_list_field",
			input: `
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
    internal.config.kubernetes.io/index: '0'
---
kind: Service
spec:
  selectors:
    foo: bar
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/index: '1'
`,
			instance: kio.ByteReadWriter{KeepReaderAnnotations: true},
		},

		{
			name: "manual_override_wrap",
			input: `
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
		{
			name: "anchors_not_inflated",
			input: `
kind: ConfigMap
metadata:
  name: foo
data:
  color: &color-used blue
  feeling: *color-used
`,
			// If YAML anchors were automagically inflated,
			// the expectedOutput would be something like
			//
			// kind: ConfigMap
			// metadata:
			//   name: foo
			// data:
			//   color: blue
			//   feeling: blue
			expectedOutput: `
kind: ConfigMap
metadata:
  name: foo
data:
  color: &color-used blue
  feeling: *color-used
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

			nodes, err := w.Read()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = w.Write(nodes)

			if tc.err != "" {
				if !assert.EqualError(t, err, tc.err) {
					t.FailNow()
				}
				return
			} else if !assert.NoError(t, err) {
				t.FailNow()
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
			name: "round_trip with mixed indentations in same resource, wide wins as it is first",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
env:
- foo
- bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
  - foo
env:
  - foo
  - bar
`,
		},
		{
			name: "round_trip with mixed indentations in same resource, compact wins as it is first",
			input: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
env:
  - foo
  - bar
`,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
spec:
- foo
env:
- foo
- bar
`,
		},
		{
			name: "unwrap ResourceList with annotations",
			input: `
apiVersion: config.kubernetes.io/v1
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

func TestByteReadWriter_WrapBareSeqNode(t *testing.T) {
	type testCase struct {
		name            string
		readerErr       string
		writerErr       string
		input           string
		wrapBareSeqNode bool
		expectedOutput  string
		instance        kio.ByteReadWriter
	}

	testCases := []testCase{
		{
			name:            "round_trip bare seq node simple",
			wrapBareSeqNode: true,
			input: `
- foo
- bar
`,
			expectedOutput: `
- foo
- bar
`,
		},
		{
			name:            "round_trip bare seq node",
			wrapBareSeqNode: true,
			input: `# Use the old CRD because of the quantity validation issue:
# https://github.com/kubeflow/kubeflow/issues/5722
- op: replace
  path: /spec
  value:
    group: kubeflow.org
    names:
      kind: Notebook
      plural: notebooks
      singular: notebook
    scope: Namespaced
    subresources:
      status: {}
    versions:
    - name: v1alpha1
      served: true
      storage: false
`,
			expectedOutput: `# Use the old CRD because of the quantity validation issue:
# https://github.com/kubeflow/kubeflow/issues/5722
- op: replace
  path: /spec
  value:
    group: kubeflow.org
    names:
      kind: Notebook
      plural: notebooks
      singular: notebook
    scope: Namespaced
    subresources:
      status: {}
    versions:
    - name: v1alpha1
      served: true
      storage: false
`,
		},
		{
			name:            "error round_trip bare seq node simple",
			wrapBareSeqNode: false,
			input: `
- foo
- bar
`,
			readerErr: "wrong Node Kind for  expected: MappingNode was SequenceNode",
		},
		{
			name:            "error round_trip bare seq node",
			wrapBareSeqNode: false,
			input: `# Use the old CRD because of the quantity validation issue:
# https://github.com/kubeflow/kubeflow/issues/5722
- op: replace
  path: /spec
  value:
    group: kubeflow.org
    names:
      kind: Notebook
      plural: notebooks
      singular: notebook
    scope: Namespaced
    subresources:
      status: {}
    versions:
    - name: v1alpha1
      served: true
      storage: false
`,
			readerErr: "wrong Node Kind for  expected: MappingNode was SequenceNode",
		},
		{
			name:            "round_trip bare seq node json",
			wrapBareSeqNode: true,
			input:           `[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--namespaced"}]`,
			expectedOutput:  `[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--namespaced"}]`,
		},
		{
			name:            "error round_trip invalid yaml node",
			wrapBareSeqNode: false,
			input:           "I am not valid",
			readerErr:       "wrong Node Kind for  expected: MappingNode was ScalarNode",
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
			w.WrapBareSeqNode = tc.wrapBareSeqNode

			nodes, err := w.Read()
			if tc.readerErr != "" {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), tc.readerErr) {
					t.FailNow()
				}
				return
			}

			w.WrappingKind = ""
			err = w.Write(nodes)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if tc.writerErr != "" {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), tc.writerErr) {
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

func TestByteReadWriter_ResourceListWrapping(t *testing.T) {
	singleDeployment := `kind: Deployment
apiVersion: v1
metadata:
  name: tester
  namespace: default
spec:
  replicas: 0`
	resourceList := `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- kind: Deployment
  apiVersion: v1
  metadata:
    name: tester
    namespace: default
  spec:
    replicas: 0`
	resourceListWithError := `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- kind: Deployment
  apiVersion: v1
  metadata:
    name: tester
    namespace: default
  spec:
    replicas: 0
results:
- message: some error
  severity: error`
	resourceListDifferentWrapper := strings.NewReplacer(
		"kind: ResourceList", "kind: SomethingElse",
		"apiVersion: config.kubernetes.io/v1", "apiVersion: fakeVersion",
	).Replace(resourceList)

	testCases := []struct {
		desc           string
		noWrap         bool
		wrapKind       string
		wrapAPIVersion string
		input          string
		want           string
	}{
		{
			desc:  "resource list",
			input: resourceList,
			want:  resourceList,
		},
		{
			desc:  "individual resources",
			input: singleDeployment,
			want:  singleDeployment,
		},
		{
			desc:           "no nested wrapping",
			wrapKind:       kio.ResourceListKind,
			wrapAPIVersion: kio.ResourceListAPIVersion,
			input:          resourceList,
			want:           resourceList,
		},
		{
			desc:   "unwrap resource list",
			noWrap: true,
			input:  resourceList,
			want:   singleDeployment,
		},
		{
			desc:           "wrap individual resources",
			wrapKind:       kio.ResourceListKind,
			wrapAPIVersion: kio.ResourceListAPIVersion,
			input:          singleDeployment,
			want:           resourceList,
		},
		{
			desc:           "NoWrap has precedence",
			noWrap:         true,
			wrapKind:       kio.ResourceListKind,
			wrapAPIVersion: kio.ResourceListAPIVersion,
			input:          singleDeployment,
			want:           singleDeployment,
		},
		{
			desc:           "honor specified wrapping kind",
			wrapKind:       "SomethingElse",
			wrapAPIVersion: "fakeVersion",
			input:          resourceList,
			want:           resourceListDifferentWrapper,
		},
		{
			desc:  "passthrough results",
			input: resourceListWithError,
			want:  resourceListWithError,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.desc, func(t *testing.T) {
			var got bytes.Buffer
			rw := kio.ByteReadWriter{
				Reader:             strings.NewReader(tc.input),
				Writer:             &got,
				NoWrap:             tc.noWrap,
				WrappingAPIVersion: tc.wrapAPIVersion,
				WrappingKind:       tc.wrapKind,
			}

			rnodes, err := rw.Read()
			assert.NoError(t, err)

			err = rw.Write(rnodes)
			assert.NoError(t, err)

			assert.Equal(t, tc.want, strings.TrimSpace(got.String()))
		})
	}
}
