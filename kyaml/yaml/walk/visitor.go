// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type ListKind int32

const (
	AssociativeList ListKind = 1 + iota
	NonAssociateList
)

// Visitor is invoked by walk with source and destination node pairs
type Visitor interface {
	VisitMap(nodes Sources) (*yaml.RNode, error)

	VisitScalar(nodes Sources) (*yaml.RNode, error)

	VisitList(nodes Sources, kind ListKind) (*yaml.RNode, error)
}

// ClearNode is returned if GrepFilter should do nothing after calling Set
var ClearNode *yaml.RNode
