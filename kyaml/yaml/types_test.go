// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		APIVersion: "rbac.istio.io/v1alpha1",
		Kind:       "ServiceRole",
		ObjectMeta: ObjectMeta{
			Name:        "wildcard",
			Namespace:   "default",
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

func TestIsYNodeTaggedNull(t *testing.T) {
	if IsYNodeTaggedNull(nil) {
		t.Fatalf("nil cannot be tagged null")
	}
	if IsYNodeTaggedNull(&Node{}) {
		t.Fatalf("untagged node is not tagged")
	}
	if IsYNodeTaggedNull(&Node{Tag: NodeTagFloat}) {
		t.Fatalf("float tagged node is not tagged")
	}
	if !IsYNodeTaggedNull(&Node{Tag: NodeTagNull}) {
		t.Fatalf("tagged node is tagged")
	}
}

func TestIsYNodeEmptyMap(t *testing.T) {
	if IsYNodeEmptyMap(nil) {
		t.Fatalf("nil cannot be a map")
	}
	if IsYNodeEmptyMap(&Node{}) {
		t.Fatalf("raw node is not a map")
	}
	if IsYNodeEmptyMap(&Node{Kind: SequenceNode}) {
		t.Fatalf("seq node is not a map")
	}
	n := &Node{Kind: MappingNode}
	if !IsYNodeEmptyMap(n) {
		t.Fatalf("empty mapping node is an empty mapping node")
	}
	n.Content = append(n.Content, &Node{Kind: SequenceNode})
	if IsYNodeEmptyMap(n) {
		t.Fatalf("a node with content isn't empty")
	}
}

func TestIsYNodeEmptySeq(t *testing.T) {
	if IsYNodeEmptySeq(nil) {
		t.Fatalf("nil cannot be a map")
	}
	if IsYNodeEmptySeq(&Node{}) {
		t.Fatalf("raw node is not a map")
	}
	if IsYNodeEmptySeq(&Node{Kind: MappingNode}) {
		t.Fatalf("map node is not a sequence")
	}
	n := &Node{Kind: SequenceNode}
	if !IsYNodeEmptySeq(n) {
		t.Fatalf("empty sequence node is an empty sequence node")
	}
	n.Content = append(n.Content, &Node{Kind: MappingNode})
	if IsYNodeEmptySeq(n) {
		t.Fatalf("a node with content isn't empty")
	}
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

func TestMapNodeIsNilOrEmpty(t *testing.T) {
	var mn *MapNode

	if !mn.IsNilOrEmpty() {
		t.Fatalf("nil should be empty")
	}

	mn = &MapNode{Key: MakeNullNode()}
	if !mn.IsNilOrEmpty() {
		t.Fatalf("missing value should be empty")
	}

	mn.Value = NewRNode(nil)
	if !mn.IsNilOrEmpty() {
		t.Fatalf("missing value YNode should be empty")
	}

	mn.Value = MakeNullNode()
	if !mn.IsNilOrEmpty() {
		t.Fatalf("value tagged null should be empty")
	}

	mn.Value = NewMapRNode(nil)
	if !mn.IsNilOrEmpty() {
		t.Fatalf("empty map should be empty")
	}

	mn.Value = NewMapRNode(&map[string]string{"foo": "bar"})
	if mn.IsNilOrEmpty() {
		t.Fatalf("non-empty map should not be empty")
	}

	mn.Value = NewListRNode()
	if !mn.IsNilOrEmpty() {
		t.Fatalf("empty list should be empty")
	}

	mn.Value = NewListRNode("foo")
	if mn.IsNilOrEmpty() {
		t.Fatalf("non-empty list should not be empty")
	}
}
