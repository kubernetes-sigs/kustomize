// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter walks the Source RNode and modifies the RNode provided to GrepFilter.
type Walker struct {
	// Visitor is invoked by GrepFilter
	Visitor

	// Source is the RNode to walk.  All Source fields and associative list elements
	// will be visited.
	Sources Sources

	// Path is the field path to the current Source Node.
	Path []string
}

func (l Walker) Kind() yaml.Kind {
	for _, s := range l.Sources {
		if !yaml.IsEmpty(s) {
			return s.YNode().Kind
		}
	}
	return 0
}

// GrepFilter implements yaml.GrepFilter
func (l Walker) Walk() (*yaml.RNode, error) {
	// invoke the handler for the corresponding node type
	switch l.Kind() {
	case yaml.MappingNode:
		if err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.MappingNode, l.Sources...); err != nil {
			return nil, err
		}
		return l.walkMap()
	case yaml.SequenceNode:
		if err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.SequenceNode, l.Sources...); err != nil {
			return nil, err
		}
		if yaml.IsAssociative(l.Sources) {
			return l.walkAssociativeSequence()
		}
		return l.walkNonAssociativeSequence()

	case yaml.ScalarNode:
		if err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.ScalarNode, l.Sources...); err != nil {
			return nil, err
		}
		return l.walkScalar()
	case 0:
		// walk empty nodes as maps
		return l.walkMap()
	default:
		return nil, nil
	}
}

const (
	DestIndex = iota
	OriginIndex
	UpdatedIndex
)

type Sources []*yaml.RNode

// Dest returns the destination node
func (s Sources) Dest() *yaml.RNode {
	if len(s) <= DestIndex {
		return nil
	}
	return s[DestIndex]
}

// Origin returns the origin node
func (s Sources) Origin() *yaml.RNode {
	if len(s) <= OriginIndex {
		return nil
	}
	return s[OriginIndex]
}

// Updated returns the updated node
func (s Sources) Updated() *yaml.RNode {
	if len(s) <= UpdatedIndex {
		return nil
	}
	return s[UpdatedIndex]
}

func (s Sources) String() string {
	var values []string
	for i := range s {
		str, err := s[i].String()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		values = append(values, str)
	}
	return strings.Join(values, "\n")
}

// setDestNode sets the destination source node
func (s Sources) setDestNode(node *yaml.RNode, err error) (*yaml.RNode, error) {
	if err != nil {
		return nil, err
	}
	s[0] = node
	return node, nil
}

type FieldSources []*yaml.MapNode

// Dest returns the destination node
func (s FieldSources) Dest() *yaml.MapNode {
	if len(s) <= DestIndex {
		return nil
	}
	return s[DestIndex]
}

// Origin returns the origin node
func (s FieldSources) Origin() *yaml.MapNode {
	if len(s) <= OriginIndex {
		return nil
	}
	return s[OriginIndex]
}

// Updated returns the updated node
func (s FieldSources) Updated() *yaml.MapNode {
	if len(s) <= UpdatedIndex {
		return nil
	}
	return s[UpdatedIndex]
}
