// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package comments

import (
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
)

// CopyComments recursively copies the comments on fields in from to fields in to
func CopyComments(from, to *yaml.RNode) error {
	// walk the fields copying comments
	_, err := walk.Walker{
		Sources:            []*yaml.RNode{from, to},
		Visitor:            &copier{},
		VisitKeysAsScalars: true}.Walk()
	return err
}

// copier implements walk.Visitor, and copies comments to fields shared between 2 instances
// of a resource
type copier struct{}

func (c *copier) VisitMap(s walk.Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	copy(s.Dest(), s.Origin())
	return s.Dest(), nil
}

func (c *copier) VisitScalar(s walk.Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	copy(s.Dest(), s.Origin())
	return s.Dest(), nil
}

func (c *copier) VisitList(s walk.Sources, _ *openapi.ResourceSchema, _ walk.ListKind) (
	*yaml.RNode, error) {
	copy(s.Dest(), s.Origin())
	return s.Dest(), nil
}

// copy copies the comment from one field to another
func copy(from, to *yaml.RNode) {
	if from == nil || to == nil {
		return
	}
	if from.YNode().LineComment != "" {
		to.YNode().LineComment = from.YNode().LineComment
	}
	if from.YNode().HeadComment != "" {
		to.YNode().HeadComment = from.YNode().HeadComment
	}
	if from.YNode().FootComment != "" {
		to.YNode().FootComment = from.YNode().FootComment
	}
}
