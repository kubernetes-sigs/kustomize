// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"reflect"
	"testing"
)

func TestCopyYNode(t *testing.T) {
	ynSub1 := Node{
		Kind: 100,
	}
	ynSub2 := Node{
		Kind: 200,
	}
	ynSub3 := Node{
		Kind: 300,
	}
	yn := Node{
		Kind:        5000,
		Style:       6000,
		Tag:         "red",
		Value:       "green",
		Anchor:      "blue",
		Alias:       &ynSub3,
		Content:     []*Node{&ynSub1, &ynSub2},
		HeadComment: "apple",
		LineComment: "peach",
		FootComment: "banana",
		Line:        7000,
		Column:      8000,
	}
	ynAddr := &yn
	if !reflect.DeepEqual(&yn, ynAddr) {
		t.Fatalf("truly %v should equal %v", &yn, ynAddr)
	}
	ynC := CopyYNode(&yn)
	if !reflect.DeepEqual(yn.Content, ynC.Content) {
		t.Fatalf("copy content %v is not deep equal to %v", ynC, yn)
	}
	if !reflect.DeepEqual(&yn, ynC) {
		t.Fatalf("\noriginal: %v\n    copy: %v\nShould be equal.", yn, ynC)
	}
	tmp := yn.Content[0].Kind
	yn.Content[0].Kind = 666
	if reflect.DeepEqual(&yn, ynC) {
		t.Fatalf("changing component should break equality")
	}
	yn.Content[0].Kind = tmp
	if !reflect.DeepEqual(&yn, ynC) {
		t.Fatalf("should be okay now")
	}
	yn.Tag = "Different"
	if yn.Tag == ynC.Tag {
		t.Fatalf("field aliased!")
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
