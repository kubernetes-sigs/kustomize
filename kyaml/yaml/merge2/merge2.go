// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package merge contains libraries for merging fields from one RNode to another
// RNode
package merge2

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
)

// Merge merges fields from src into dest.
func Merge(src, dest *yaml.RNode) (*yaml.RNode, error) {
	return walk.Walker{Sources: []*yaml.RNode{dest, src}, Visitor: Merger{}}.Walk()
}

// Merge parses the arguments, and merges fields from srcStr into destStr.
func MergeStrings(srcStr, destStr string) (string, error) {
	src, err := yaml.Parse(srcStr)
	if err != nil {
		return "", err
	}
	dest, err := yaml.Parse(destStr)
	if err != nil {
		return "", err
	}

	result, err := Merge(src, dest)
	if err != nil {
		return "", err
	}
	return result.String()
}

type Merger struct {
	// for forwards compatibility when new functions are added to the interface
}

var _ walk.Visitor = Merger{}

func (m Merger) VisitMap(nodes walk.Sources) (*yaml.RNode, error) {
	if err := m.SetComments(nodes); err != nil {
		return nil, err
	}
	if yaml.IsEmpty(nodes.Dest()) {
		// Add
		return nodes.Origin(), nil
	}
	if yaml.IsNull(nodes.Origin()) {
		// clear the value
		return walk.ClearNode, nil
	}
	// Recursively Merge dest
	return nodes.Dest(), nil
}

func (m Merger) VisitScalar(nodes walk.Sources) (*yaml.RNode, error) {
	if err := m.SetComments(nodes); err != nil {
		return nil, err
	}
	// Override value
	if nodes.Origin() != nil {
		return nodes.Origin(), nil
	}
	// Keep
	return nodes.Dest(), nil
}

func (m Merger) VisitList(nodes walk.Sources, kind walk.ListKind) (*yaml.RNode, error) {
	if err := m.SetComments(nodes); err != nil {
		return nil, err
	}
	if kind == walk.NonAssociateList {
		// Override value
		if nodes.Origin() != nil {
			return nodes.Origin(), nil
		}
		// Keep
		return nodes.Dest(), nil
	}

	// Add
	if yaml.IsEmpty(nodes.Dest()) {
		return nodes.Origin(), nil
	}
	// Clear
	if yaml.IsNull(nodes.Origin()) {
		return walk.ClearNode, nil
	}
	// Recursively Merge dest
	return nodes.Dest(), nil
}

// SetComments copies the dest comments to the source comments if they are present
// on the source.
func (m Merger) SetComments(sources walk.Sources) error {
	source := sources.Origin()
	dest := sources.Dest()
	if source != nil && source.YNode().FootComment != "" {
		dest.YNode().FootComment = source.YNode().FootComment
	}
	if source != nil && source.YNode().HeadComment != "" {
		dest.YNode().HeadComment = source.YNode().HeadComment
	}
	if source != nil && source.YNode().LineComment != "" {
		dest.YNode().LineComment = source.YNode().LineComment
	}
	return nil
}
