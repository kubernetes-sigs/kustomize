// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/schema"
)

// Walker walks the Source RNode and modifies the RNode provided to GrepFilter.
type Walker struct {
	// Visitor is invoked by GrepFilter
	Visitor

	Schema *openapi.ResourceSchema

	// Source is the RNode to walk.  All Source fields and associative list elements
	// will be visited.
	Sources Sources

	// Path is the field path to the current Source Node.
	Path []string

	// InferAssociativeLists if set to true will infer merge strategies for
	// fields which it doesn't have the schema based on the fields in the
	// list elements.
	InferAssociativeLists bool

	// VisitKeysAsScalars if true will call VisitScalar on map entry keys,
	// providing nil as the OpenAPI schema.
	VisitKeysAsScalars bool

	// MergeOptions is a struct to store options for merge
	MergeOptions yaml.MergeOptions
}

// Kind returns the kind of the first non-null node in Sources.
func (l Walker) Kind() yaml.Kind {
	for _, s := range l.Sources {
		if !yaml.IsMissingOrNull(s) {
			return s.YNode().Kind
		}
	}
	return 0
}

// Walk will recursively traverse every item in the Sources and perform corresponding
// actions on them
func (l Walker) Walk() (*yaml.RNode, error) {
	l.Schema = l.GetSchema()

	// invoke the handler for the corresponding node type
	switch l.Kind() {
	case yaml.MappingNode:
		if err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.MappingNode, l.Sources...); err != nil {
			// If AllowKindChange is set and there's a type mismatch, allow the
			// origin (source) to replace dest entirely instead of erroring.
			if l.MergeOptions.AllowKindChange {
				if replaced := l.replaceOnKindMismatch(); replaced != nil {
					return replaced, nil
				}
			}
			return nil, err
		}
		return l.walkMap()
	case yaml.SequenceNode:
		if err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.SequenceNode, l.Sources...); err != nil {
			// If AllowKindChange is set and there's a type mismatch, allow the
			// origin (source) to replace dest entirely instead of erroring.
			if l.MergeOptions.AllowKindChange {
				if replaced := l.replaceOnKindMismatch(); replaced != nil {
					return replaced, nil
				}
			}
			return nil, err
		}
		// AssociativeSequence means the items in the sequence are associative. They can be merged
		// according to merge key.
		if schema.IsAssociative(l.Schema, l.Sources, l.InferAssociativeLists) {
			return l.walkAssociativeSequence()
		}
		return l.walkNonAssociativeSequence()

	case yaml.ScalarNode:
		if err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.ScalarNode, l.Sources...); err != nil {
			// If AllowKindChange is set and there's a type mismatch, allow the
			// origin (source) to replace dest entirely instead of erroring.
			if l.MergeOptions.AllowKindChange {
				if replaced := l.replaceOnKindMismatch(); replaced != nil {
					return replaced, nil
				}
			}
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

// replaceOnKindMismatch returns the origin node when there's a kind mismatch
// between dest and origin. This allows the origin to completely replace the
// dest when they have different types (e.g., map vs list, map vs scalar).
// Returns nil if no replacement should occur.
func (l Walker) replaceOnKindMismatch() *yaml.RNode {
	dest := l.Sources.Dest()
	origin := l.Sources.Origin()

	// If origin is nil or null, we can't do a replacement
	if yaml.IsMissingOrNull(origin) {
		return nil
	}

	// If dest is nil or null, origin should be used directly
	if yaml.IsMissingOrNull(dest) {
		return origin
	}

	// Check if origin has a different kind than dest - if so, origin replaces dest
	if origin.YNode().Kind != dest.YNode().Kind {
		return origin
	}

	return nil
}

func (l Walker) GetSchema() *openapi.ResourceSchema {
	for i := range l.Sources {
		r := l.Sources[i]
		if yaml.IsMissingOrNull(r) {
			continue
		}

		fm := fieldmeta.FieldMeta{}
		if err := fm.Read(r); err == nil && !fm.IsEmpty() {
			// per-field schema, this is fine
			if fm.Schema.Ref.String() != "" {
				// resolve the reference
				s, err := openapi.Resolve(&fm.Schema.Ref, openapi.Schema())
				if err == nil && s != nil {
					fm.Schema = *s
				}
			}
			return &openapi.ResourceSchema{Schema: &fm.Schema}
		}
	}

	if l.Schema != nil {
		return l.Schema
	}
	for i := range l.Sources {
		r := l.Sources[i]
		if yaml.IsMissingOrNull(r) {
			continue
		}

		m, _ := r.GetMeta()
		if m.Kind == "" || m.APIVersion == "" {
			continue
		}

		s := openapi.SchemaForResourceType(yaml.TypeMeta{Kind: m.Kind, APIVersion: m.APIVersion})
		if s != nil {
			return s
		}
	}
	return nil
}

const (
	DestIndex = iota
	OriginIndex
	UpdatedIndex
)

// Sources are a list of RNodes. First item is the dest node, followed by
// multiple source nodes.
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
