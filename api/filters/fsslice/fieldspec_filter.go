// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fsslice

import (
	"strings"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// fieldSpecFilter applies a single fieldSpec to a single object
// fieldSpecFilter stores internal state and should not be reused
type fieldSpecFilter struct {
	// FieldSpec contains the path to the value to set.
	FieldSpec types.FieldSpec `yaml:"fieldSpec"`

	// Set the field using this function
	SetValue SetFn

	// CreateKind defines the type of node to create if the field is not found
	CreateKind yaml.Kind

	CreateTag string

	// path keeps internal state about the current path
	path []string
}

func (fltr fieldSpecFilter) Filter(obj *yaml.RNode) (*yaml.RNode, error) {
	// check if the FieldSpec applies to the object
	if match, err := isMatchGVK(fltr.FieldSpec, obj); !match || err != nil {
		return obj, errors.Wrap(err)
	}
	fltr.path = strings.Split(fltr.FieldSpec.Path, "/")
	if err := fltr.filter(obj); err != nil {
		s, _ := obj.String()
		return nil, errors.WrapPrefixf(err,
			"obj %v at path %v", s, fltr.FieldSpec.Path)
	}
	return obj, nil
}

func (fltr fieldSpecFilter) filter(obj *yaml.RNode) error {
	if len(fltr.path) == 0 {
		// found the field -- set its value
		return fltr.SetValue(obj)
	}
	switch obj.YNode().Kind {
	case yaml.SequenceNode:
		return fltr.seq(obj)
	case yaml.MappingNode:
		return fltr.field(obj)
	}
	// not found -- this might be an error since the type doesn't match

	return errors.Errorf("unsupported yaml node")
}

// field calls filter on the field matching the next path element
func (fltr fieldSpecFilter) field(obj *yaml.RNode) error {
	fieldName, isSeq := isSequenceField(fltr.path[0])

	// lookup the field matching the next path element
	var lookupField yaml.Filter
	var kind yaml.Kind
	var tag string
	switch {
	case !fltr.FieldSpec.CreateIfNotPresent || fltr.CreateKind == 0 || isSeq:
		// dont' create the field if we don't find it
		lookupField = yaml.Lookup(fieldName)
	case len(fltr.path) <= 1:
		// create the field if it is missing: use the provided node kind
		lookupField = yaml.LookupCreate(fltr.CreateKind, fieldName)
		kind = fltr.CreateKind
		tag = fltr.CreateTag
	default:
		// create the field if it is missing: must be a mapping node
		lookupField = yaml.LookupCreate(yaml.MappingNode, fieldName)
		kind = yaml.MappingNode
		tag = "!!map"
	}

	// locate (or maybe create) the field
	field, err := obj.Pipe(lookupField)
	if err != nil || field == nil {
		return errors.WrapPrefixf(err, "fieldName: %s", fieldName)
	}

	// if the value exists, but is null, then change it to the creation type
	// TODO: update yaml.LookupCreate to support this
	if field.YNode().Tag == "!!null" {
		field.YNode().Kind = kind
		field.YNode().Tag = tag
	}

	// copy the current fltr and change the path on the copy
	var next = fltr
	// call filter for the next path element on the matching field
	next.path = fltr.path[1:]
	return next.filter(field)
}

// seq calls filter on all sequence elements
func (fltr fieldSpecFilter) seq(obj *yaml.RNode) error {
	if err := obj.VisitElements(func(node *yaml.RNode) error {
		// recurse on each element -- re-allocating a fieldSpecFilter is
		// not strictly required, but is more consistent with field
		// and less likely to have side effects
		// keep the entire path -- it does not contain parts for sequences
		return fltr.filter(node)
	}); err != nil {
		return errors.WrapPrefixf(err,
			"visit traversal on path: %v", fltr.path)
	}

	return nil
}

// isSequenceField returns true if the path element is for a sequence field.
// isSequence also returns the path element with the '[]' suffix trimmed
func isSequenceField(name string) (string, bool) {
	isSeq := strings.HasSuffix(name, "[]")
	name = strings.TrimSuffix(name, "[]")
	return name, isSeq
}

// isMatchGVK returns true if the fs.GVK matches the obj GVK.
func isMatchGVK(fs types.FieldSpec, obj *yaml.RNode) (bool, error) {
	meta, err := obj.GetMeta()
	if err != nil {
		return false, err
	}
	if fs.Kind != "" && fs.Kind != meta.Kind {
		// kind doesn't match
		return false, err
	}

	// parse the group and version from the apiVersion field
	group, version := parseGV(meta.APIVersion)

	if fs.Group != "" && fs.Group != group {
		// group doesn't match
		return false, nil
	}

	if fs.Version != "" && fs.Version != version {
		// version doesn't match
		return false, nil
	}

	return true, nil
}
