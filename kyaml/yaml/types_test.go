// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
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
