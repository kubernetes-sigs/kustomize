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
