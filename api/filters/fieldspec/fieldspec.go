// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"strings"

	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = Filter{}

// Filter applies a single fieldSpec to a single object
// Filter stores internal state and should not be reused
type Filter struct {
	// FieldSpec contains the path to the value to set.
	FieldSpec types.FieldSpec `yaml:"fieldSpec"`

	// Set the field using this function
	SetValue filtersutil.SetFn

	// CreateKind defines the type of node to create if the field is not found
	CreateKind yaml.Kind

	CreateTag string

	// path keeps internal state about the current path
	path []string
}

func (fltr Filter) Filter(obj *yaml.RNode) (*yaml.RNode, error) {
	// check if the FieldSpec applies to the object
	if match, err := isMatchGVK(fltr.FieldSpec, obj); !match || err != nil {
		return obj, errors.Wrap(err)
	}
	fltr.path = splitPath(fltr.FieldSpec.Path)
	if err := fltr.filter(obj); err != nil {
		s, _ := obj.String()
		return nil, errors.WrapPrefixf(err,
			"obj '%s' at path '%v'", s, fltr.FieldSpec.Path)
	}
	return obj, nil
}

func (fltr Filter) filter(obj *yaml.RNode) error {
	if len(fltr.path) == 0 {
		// found the field -- set its value
		return fltr.SetValue(obj)
	}
	if obj.IsTaggedNull() {
		return nil
	}
	switch obj.YNode().Kind {
	case yaml.SequenceNode:
		return fltr.seq(obj)
	case yaml.MappingNode:
		return fltr.field(obj)
	default:
		return errors.Errorf("expected sequence or mapping node")
	}
}

// field calls filter on the field matching the next path element
func (fltr Filter) field(obj *yaml.RNode) error {
	fieldName, isSeq := isSequenceField(fltr.path[0])
	// lookup the field matching the next path element
	var lookupField yaml.Filter
	var kind yaml.Kind
	tag := yaml.NodeTagEmpty
	switch {
	case !fltr.FieldSpec.CreateIfNotPresent || fltr.CreateKind == 0 || isSeq:
		// dont' create the field if we don't find it
		lookupField = yaml.Lookup(fieldName)
		if isSeq {
			// The query path thinks this field should be a sequence;
			// accept this hint for use later if the tag is NodeTagNull.
			kind = yaml.SequenceNode
		}
	case len(fltr.path) <= 1:
		// create the field if it is missing: use the provided node kind
		lookupField = yaml.LookupCreate(fltr.CreateKind, fieldName)
		kind = fltr.CreateKind
		tag = fltr.CreateTag
	default:
		// create the field if it is missing: must be a mapping node
		lookupField = yaml.LookupCreate(yaml.MappingNode, fieldName)
		kind = yaml.MappingNode
		tag = yaml.NodeTagMap
	}

	// locate (or maybe create) the field
	field, err := obj.Pipe(lookupField)
	if err != nil || field == nil {
		return errors.WrapPrefixf(err, "fieldName: %s", fieldName)
	}

	// if the value exists, but is null and kind is set,
	// then change it to the creation type
	// TODO: update yaml.LookupCreate to support this
	if field.YNode().Tag == yaml.NodeTagNull && yaml.IsCreate(kind) {
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
func (fltr Filter) seq(obj *yaml.RNode) error {
	if err := obj.VisitElements(func(node *yaml.RNode) error {
		// recurse on each element -- re-allocating a Filter is
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

func splitPath(path string) []string {
	ps := strings.Split(path, "/")
	var res []string
	res = append(res, ps[0])
	for i := 1; i < len(ps); i++ {
		lastIndex := len(res) - 1
		if strings.HasSuffix(res[lastIndex], "\\") {
			res[lastIndex] = strings.TrimSuffix(res[lastIndex], "\\") + "/" + ps[i]
		} else {
			res = append(res, ps[i])
		}
	}
	return res
}
