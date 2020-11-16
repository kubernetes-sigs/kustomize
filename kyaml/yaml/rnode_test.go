// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRNodeHasNilEntryInList(t *testing.T) {
	testConfigMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	}
	type resultExpected struct {
		hasNil bool
		path   string
	}

	testCases := map[string]struct {
		theMap map[string]interface{}
		rsExp  resultExpected
	}{
		"actuallyNil": {
			theMap: nil,
			rsExp:  resultExpected{},
		},
		"empty": {
			theMap: map[string]interface{}{},
			rsExp:  resultExpected{},
		},
		"list": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{},
		},
		"listWithNil": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					nil,
				},
			},
			rsExp: resultExpected{
				hasNil: false, // TODO: This should be true.
				path:   "this/should/be/non-empty",
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			hasNil, path := rn.HasNilEntryInList()
			if tc.rsExp.hasNil {
				if !assert.True(t, hasNil) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.rsExp.path, path) {
					t.FailNow()
				}
			} else {
				if !assert.False(t, hasNil) {
					t.FailNow()
				}
				if !assert.Empty(t, path) {
					t.FailNow()
				}
			}
		})
	}
}

func TestRNodeGetValidatedMetadata(t *testing.T) {
	testConfigMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	}
	type resultExpected struct {
		out    ResourceMeta
		errMsg string
	}

	testCases := map[string]struct {
		theMap map[string]interface{}
		rsExp  resultExpected
	}{
		"actuallyNil": {
			theMap: nil,
			rsExp: resultExpected{
				errMsg: "missing Resource metadata",
			},
		},
		"empty": {
			theMap: map[string]interface{}{},
			rsExp: resultExpected{
				errMsg: "missing Resource metadata",
			},
		},
		"mostlyEmpty": {
			theMap: map[string]interface{}{
				"hey": "there",
			},
			rsExp: resultExpected{
				errMsg: "missing Resource metadata",
			},
		},
		"noNameConfigMap": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
			},
			rsExp: resultExpected{
				errMsg: "missing metadata.name",
			},
		},
		"configmap": {
			theMap: testConfigMap,
			rsExp: resultExpected{
				out: ResourceMeta{
					TypeMeta: TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: ObjectMeta{
						NameMeta: NameMeta{
							Name: "winnie",
						},
					},
				},
			},
		},
		"list": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: ResourceMeta{
					TypeMeta: TypeMeta{
						APIVersion: "v1",
						Kind:       "List",
					},
				},
			},
		},
		"configmaplist": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMapList",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: ResourceMeta{
					TypeMeta: TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMapList",
					},
				},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			m, err := rn.GetValidatedMetadata()
			if tc.rsExp.errMsg == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.rsExp.out, m) {
					t.FailNow()
				}
			} else {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), tc.rsExp.errMsg) {
					t.FailNow()
				}
			}
		})
	}
}

func TestRNodeFromMap(t *testing.T) {
	testConfigMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	}
	type resultExpected struct {
		out string
		err error
	}

	testCases := map[string]struct {
		theMap map[string]interface{}
		rsExp  resultExpected
	}{
		"actuallyNil": {
			theMap: nil,
			rsExp: resultExpected{
				out: `{}`,
				err: nil,
			},
		},
		"empty": {
			theMap: map[string]interface{}{},
			rsExp: resultExpected{
				out: `{}`,
				err: nil,
			},
		},
		"mostlyEmpty": {
			theMap: map[string]interface{}{
				"hey": "there",
			},
			rsExp: resultExpected{
				out: `hey: there`,
				err: nil,
			},
		},
		"configmap": {
			theMap: testConfigMap,
			rsExp: resultExpected{
				out: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`,
				err: nil,
			},
		},
		"list": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: `
apiVersion: v1
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
kind: List
`,
				err: nil,
			},
		},
		"configmaplist": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMapList",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: `
apiVersion: v1
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
kind: ConfigMapList
`,
				err: nil,
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if tc.rsExp.err == nil {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tc.rsExp.out),
					strings.TrimSpace(rn.MustString())) {
					t.FailNow()
				}
			} else {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.rsExp.err, err) {
					t.FailNow()
				}
			}
		})
	}
}

// Test that non-UTF8 characters in comments don't cause failures
func TestRNode_GetMeta_UTF16(t *testing.T) {
	sr, err := Parse(`apiVersion: rbac.istio.io/v1alpha1
kind: ServiceRole
metadata:
  name: wildcard
  namespace: default
  # If set to [“*”], it refers to all services in the namespace
  annotations:
    foo: bar
spec:
  rules:
    # There is one service in default namespace, should not result in a validation error
    # If set to [“*”], it refers to all services in the namespace
    - services: ["*"]
      methods: ["GET", "HEAD"]
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	actual, err := sr.GetMeta()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := ResourceMeta{
		TypeMeta: TypeMeta{
			APIVersion: "rbac.istio.io/v1alpha1",
			Kind:       "ServiceRole",
		},
		ObjectMeta: ObjectMeta{
			NameMeta: NameMeta{
				Name:      "wildcard",
				Namespace: "default",
			},
			Annotations: map[string]string{"foo": "bar"},
		},
	}
	if !assert.Equal(t, expected, actual) {
		t.FailNow()
	}
}

func TestRNode_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		output   string
	}{
		{
			testName: "simple document",
			input:    `{"hello":"world"}`,
			output:   `hello: world`,
		},
		{
			testName: "nested structure",
			input: `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "my-deployment",
    "namespace": "default"
  }
}
`,
			output: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			instance := &RNode{}
			err := instance.UnmarshalJSON([]byte(tc.input))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := instance.String()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tc.output), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}

func TestRNode_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		ydoc string
		want string
	}{
		{
			name: "object",
			ydoc: `
hello: world
`,
			want: `{"hello":"world"}`,
		},
		{
			name: "array",
			ydoc: `
- name: s1
- name: s2
`,
			want: `[{"name":"s1"},{"name":"s2"}]`,
		},
	}
	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			instance, err := Parse(tt.ydoc)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := instance.MarshalJSON()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tt.want), strings.TrimSpace(string(actual))) {
				t.FailNow()
			}
		})
	}
}

func TestConvertJSONToYamlNode(t *testing.T) {
	inputJSON := `{"type": "string", "maxLength": 15, "enum": ["allowedValue1", "allowedValue2"]}`
	expected := `enum:
- allowedValue1
- allowedValue2
maxLength: 15
type: string
`

	node, err := ConvertJSONToYamlNode(inputJSON)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	actual, err := node.String()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, expected, actual)
}

func TestIsMissingOrNull(t *testing.T) {
	if !IsMissingOrNull(nil) {
		t.Fatalf("input: nil")
	}
	// missing value or null value
	if !IsMissingOrNull(NewRNode(nil)) {
		t.Fatalf("input: nil value")
	}

	if IsMissingOrNull(NewScalarRNode("foo")) {
		t.Fatalf("input: valid node")
	}
	// node with NullNodeTag
	if !IsMissingOrNull(MakeNullNode()) {
		t.Fatalf("input: with NullNodeTag")
	}

	// empty array. empty array is not expected as empty
	if IsMissingOrNull(NewListRNode()) {
		t.Fatalf("input: empty array")
	}

	// array with 1 item
	node := NewListRNode("foo")
	if IsMissingOrNull(node) {
		t.Fatalf("input: array with 1 item")
	}

	// delete the item in array
	node.value.Content = nil
	if IsMissingOrNull(node) {
		t.Fatalf("input: empty array")
	}
}

func TestCopy(t *testing.T) {
	rn := RNode{
		fieldPath: []string{"fp1", "fp2"},
		value: &Node{
			Kind: 200,
		},
		Match: []string{"m1", "m2"},
	}
	rnC := rn.Copy()
	if !reflect.DeepEqual(&rn, rnC) {
		t.Fatalf("copy %v is not deep equal to %v", rnC, rn)
	}
	tmp := rn.value.Kind
	rn.value.Kind = 666
	if reflect.DeepEqual(rn, rnC) {
		t.Fatalf("changing component should break equality")
	}
	rn.value.Kind = tmp
	if !reflect.DeepEqual(&rn, rnC) {
		t.Fatalf("should be back to normal")
	}
	rn.fieldPath[0] = "Different"
	if reflect.DeepEqual(rn, rnC) {
		t.Fatalf("changing fieldpath should break equality")
	}
}

func TestFieldRNodes(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		output   []string
		err      string
	}{
		{
			testName: "simple document",
			input: `apiVersion: example.com/v1beta1
kind: Example1
spec:
  list:
  - "a"
  - "b"
  - "c"`,
			output: []string{"apiVersion", "kind", "spec", "list"},
		},
		{
			testName: "non mapping node error",
			input:    `apiVersion`,
			err:      "wrong Node Kind for  expected: MappingNode was ScalarNode: value: {apiVersion}",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			rNode, err := Parse(tc.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			fieldRNodes, err := rNode.FieldRNodes()
			if tc.err == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			} else {
				if !assert.Equal(t, tc.err, err.Error()) {
					t.FailNow()
				}
			}
			for i := range fieldRNodes {
				actual, err := fieldRNodes[i].String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.output[i], strings.TrimSpace(actual)) {
					t.FailNow()
				}
			}
		})
	}
}

func TestIsEmptyMap(t *testing.T) {
	node := NewMapRNode(nil)
	// empty map
	if !IsEmptyMap(node) {
		t.Fatalf("input: empty map")
	}
	// map with 1 item
	node = NewMapRNode(&map[string]string{
		"foo": "bar",
	})
	if IsEmptyMap(node) {
		t.Fatalf("input: map with 1 item")
	}
	// delete the item in map
	node.value.Content = nil
	if !IsEmptyMap(node) {
		t.Fatalf("input: empty map")
	}
}

func TestIsNil(t *testing.T) {
	var rn *RNode

	if !rn.IsNil() {
		t.Fatalf("uninitialized RNode should be nil")
	}

	if !NewRNode(nil).IsNil() {
		t.Fatalf("missing value YNode should be nil")
	}

	if MakeNullNode().IsNil() {
		t.Fatalf("value tagged null is not nil")
	}

	if NewMapRNode(nil).IsNil() {
		t.Fatalf("empty map not nil")
	}

	if NewListRNode().IsNil() {
		t.Fatalf("empty list not nil")
	}
}

func TestIsTaggedNull(t *testing.T) {
	var rn *RNode

	if rn.IsTaggedNull() {
		t.Fatalf("nil RNode cannot be tagged")
	}

	if NewRNode(nil).IsTaggedNull() {
		t.Fatalf("bare RNode should not be tagged")
	}

	if !MakeNullNode().IsTaggedNull() {
		t.Fatalf("a null node is tagged null by definition")
	}

	if NewMapRNode(nil).IsTaggedNull() {
		t.Fatalf("empty map should not be tagged null")
	}

	if NewListRNode().IsTaggedNull() {
		t.Fatalf("empty list should not be tagged null")
	}
}

func TestRNodeIsNilOrEmpty(t *testing.T) {
	var rn *RNode

	if !rn.IsNilOrEmpty() {
		t.Fatalf("uninitialized RNode should be empty")
	}

	if !NewRNode(nil).IsNilOrEmpty() {
		t.Fatalf("missing value YNode should be empty")
	}

	if !MakeNullNode().IsNilOrEmpty() {
		t.Fatalf("value tagged null should be empty")
	}

	if !NewMapRNode(nil).IsNilOrEmpty() {
		t.Fatalf("empty map should be empty")
	}

	if NewMapRNode(&map[string]string{"foo": "bar"}).IsNilOrEmpty() {
		t.Fatalf("non-empty map should not be empty")
	}

	if !NewListRNode().IsNilOrEmpty() {
		t.Fatalf("empty list should be empty")
	}

	if NewListRNode("foo").IsNilOrEmpty() {
		t.Fatalf("non-empty list should not be empty")
	}
}
