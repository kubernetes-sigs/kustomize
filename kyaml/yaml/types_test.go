// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"
)

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

func TestIsNameSpaceable1(t *testing.T) {
	testCases := []struct {
		tm              TypeMeta
		isNamespaceable bool
	}{
		{
			tm:              TypeMeta{Kind: "ClusterRole"},
			isNamespaceable: false,
		},
		{
			tm:              TypeMeta{Kind: "Namespace"},
			isNamespaceable: false,
		},
		{
			tm:              TypeMeta{Kind: "Deployment"},
			isNamespaceable: true,
		},
	}
	for _, tc := range testCases {
		if tc.tm.IsNamespaceable() {
			if !tc.isNamespaceable {
				t.Fatalf("%v is namespaceable, but shouldn't be", tc.tm)
			}
		} else {
			if tc.isNamespaceable {
				t.Fatalf("%v is not namespaceable, but should be", tc.tm)
			}
		}
	}
}

func TestIsNameSpaceable2(t *testing.T) {
	testCases := []struct {
		yammy           string
		isNamespaceable bool
	}{
		{
			yammy: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: bilbo
  namespace: middleEarth
`,
			isNamespaceable: true,
		},
		{
			yammy: `apiVersion: v1
kind: Namespace
metadata:
  name: middleEarth
`,
			isNamespaceable: false,
		},
	}
	for _, tc := range testCases {
		rn, err := Parse(tc.yammy)
		if err != nil {
			t.Fatalf("unexpected parse error %v from json: %s", err, tc.yammy)
		}
		meta, err := rn.GetMeta()
		if err != nil {
			t.Fatalf("unexpected meta error %v", err)
		}
		if meta.IsNamespaceable() {
			if !tc.isNamespaceable {
				t.Fatalf("%v is namespaceable, but shouldn't be", meta)
			}
		} else {
			if tc.isNamespaceable {
				t.Fatalf("%v is not namespaceable, but should be", meta)
			}
		}
	}
}
