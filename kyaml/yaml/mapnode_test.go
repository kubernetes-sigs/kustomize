// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"
)

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
