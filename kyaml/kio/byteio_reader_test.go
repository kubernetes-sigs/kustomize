// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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
			name: "wrapped_resource_list_function_config_without_items",
			input: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
functionConfig:
  foo: bar
  elems:
  - a
  - b
  - c
`,
			expectedItems: []string{},
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
    internal.config.kubernetes.io/index: '0'
`,
				`# second resource
e: f
g:
- h
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/index: '1'
`,
				`i: j
metadata:
  annotations:
    config.kubernetes.io/index: '2'
    internal.config.kubernetes.io/index: '2'
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
    internal.config.kubernetes.io/index: '0'
`,
				`
# second resource
e: f
g:
- h
metadata:
  annotations:
    config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/index: '1'
`,
				`
i: j
metadata:
  annotations:
    config.kubernetes.io/index: '2'
    internal.config.kubernetes.io/index: '2'
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
{"a": "b", "c": [1, 2], metadata: {annotations: {config.kubernetes.io/index: '0', internal.config.kubernetes.io/index: '0'}}}
`,
			},
			instance: ByteReader{},
		},

		//
		//
		//
		{
			name: "white_space_after_document_separator_should_be_ignored",
			input: `
a: b
---         
c: d
`,
			expectedItems: []string{
				`
a: b
`,
				`
c: d
`,
			},
			instance: ByteReader{OmitReaderAnnotations: true},
		},

		//
		//
		//
		{
			name: "comment_after_document_separator_should_be_ignored",
			input: `
a: b
--- #foo
c: d
`,
			expectedItems: []string{
				`
a: b
`,
				`
c: d
`,
			},
			instance: ByteReader{OmitReaderAnnotations: true},
		},

		//
		//
		//
		{
			name: "anything_after_document_separator_other_than_white_space_or_comment_is_an_error",
			input: `
a: b
--- foo
c: d
`,
			err:      "invalid document separator: --- foo",
			instance: ByteReader{OmitReaderAnnotations: true},
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

// This test shows the lower level (go-yaml) representation of a small doc
// with an anchor. The anchor structure is there, in the sense that an
// alias pointer is readily available when a node's kind is an AliasNode.
// I.e. the anchor mapping name -> object was noted during unmarshalling.
// However, at the time of writing github.com/go-yaml/yaml/encoder.go
// doesn't appear to have an option to perform anchor replacements when
// encoding.  It emits anchor definitions and references (aliases) intact.
func TestByteReader_AnchorBehavior(t *testing.T) {
	const input = `
data:
  color: &color-used blue
  feeling: *color-used
`
	expected := strings.TrimSpace(`
data:
  color: &color-used blue
  feeling: *color-used
`)
	var rNode *yaml.RNode
	{
		rNodes, err := FromBytes([]byte(input))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(rNodes))
		rNode = rNodes[0]
	}
	// Confirm internal representation.
	{
		yNode := rNode.YNode()

		// The high level object is a map of "data" to some value.
		assert.Equal(t, yaml.NodeTagMap, yNode.Tag)

		yNodes := yNode.Content
		assert.Equal(t, 2, len(yNodes))

		// Confirm that the key is "data".
		assert.Equal(t, yaml.NodeTagString, yNodes[0].Tag)
		assert.Equal(t, "data", yNodes[0].Value)

		assert.Equal(t, yaml.NodeTagMap, yNodes[1].Tag)

		// The value of the "data" key.
		yNodes = yNodes[1].Content
		// Expect two name-value pairs.
		assert.Equal(t, 4, len(yNodes))

		assert.Equal(t, yaml.ScalarNode, yNodes[0].Kind)
		assert.Equal(t, yaml.NodeTagString, yNodes[0].Tag)
		assert.Equal(t, "color", yNodes[0].Value)
		assert.Equal(t, "", yNodes[0].Anchor)
		assert.Nil(t, yNodes[0].Alias)

		assert.Equal(t, yaml.ScalarNode, yNodes[1].Kind)
		assert.Equal(t, yaml.NodeTagString, yNodes[1].Tag)
		assert.Equal(t, "blue", yNodes[1].Value)
		assert.Equal(t, "color-used", yNodes[1].Anchor)
		assert.Nil(t, yNodes[1].Alias)

		assert.Equal(t, yaml.ScalarNode, yNodes[2].Kind)
		assert.Equal(t, yaml.NodeTagString, yNodes[2].Tag)
		assert.Equal(t, "feeling", yNodes[2].Value)
		assert.Equal(t, "", yNodes[2].Anchor)
		assert.Nil(t, yNodes[2].Alias)

		assert.Equal(t, yaml.AliasNode, yNodes[3].Kind)
		assert.Equal(t, "", yNodes[3].Tag)
		assert.Equal(t, "color-used", yNodes[3].Value)
		assert.Equal(t, "", yNodes[3].Anchor)
		assert.NotNil(t, yNodes[3].Alias)
	}

	yaml, err := rNode.String()
	assert.NoError(t, err)
	assert.Equal(t, expected, strings.TrimSpace(yaml))
}

// TestByteReader_AddSeqIndentAnnotation tests if the internal.config.kubernetes.io/seqindent
// annotation is added to resources appropriately
func TestByteReader_AddSeqIndentAnnotation(t *testing.T) {
	type testCase struct {
		name                  string
		err                   string
		input                 string
		expectedAnnoValue     string
		OmitReaderAnnotations bool
	}

	testCases := []testCase{
		{
			name: "read with wide indentation",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
`,
			expectedAnnoValue: "wide",
		},
		{
			name: "read with compact indentation",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
`,
			expectedAnnoValue: "compact",
		},
		{
			name: "read with mixed indentation, wide wins",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
env:
- foo
- bar
`,
			expectedAnnoValue: "wide",
		},
		{
			name: "read with mixed indentation, compact wins",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
env:
  - foo
  - bar
`,
			expectedAnnoValue: "compact",
		},
		{
			name: "error if conflicting options",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
env:
  - foo
  - bar
`,
			OmitReaderAnnotations: true,
			err:                   `"PreserveSeqIndent" option adds a reader annotation, please set "OmitReaderAnnotations" to false`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			rNodes, err := (&ByteReader{
				OmitReaderAnnotations: tc.OmitReaderAnnotations,
				PreserveSeqIndent:     true,
				Reader:                bytes.NewBuffer([]byte(tc.input)),
			}).Read()
			if tc.err != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err.Error())
				return
			}
			assert.NoError(t, err)
			actual := rNodes[0].GetAnnotations()[kioutil.SeqIndentAnnotation]
			assert.Equal(t, tc.expectedAnnoValue, actual)
		})
	}
}
