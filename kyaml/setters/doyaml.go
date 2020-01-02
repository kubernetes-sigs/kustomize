// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package sub substitutes strings in fields
package setters

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &fieldSetter{}

// fieldSetter sets part or all of a field value.
type fieldSetter struct {
	// Name is the name of the setter to perform.
	Name string

	// Value is the value to set.
	Value string

	// Description, if specified will set 'description' for the field.  Optional.
	Description string

	// SetBy, if specified will set 'setBy' for the field.  Optional.
	SetBy string

	// Count is incremented by Filter for each field that is set.
	Count int
}

// Filter implements yaml.Filter
func (fs *fieldSetter) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		// Document is the root of the object and always contains 1 node
		return fs.Filter(yaml.NewRNode(object.YNode().Content[0]))
	case yaml.MappingNode:
		return object, object.VisitFields(func(node *yaml.MapNode) error {
			// Traverse each field value
			return node.Value.PipeE(fs)
		})
	case yaml.SequenceNode:
		return object, object.VisitElements(func(node *yaml.RNode) error {
			// Traverse each list element
			return node.PipeE(fs)
		})
	case yaml.ScalarNode:
		// Check if there is a setter matching the name
		s, f, partial, err := fs.findSetter(object)
		if err != nil {
			return nil, err
		}
		if s == nil {
			// no matching setter
			return object, nil
		}
		// set the field value
		return object, fs.set(object, s, f, partial)
	default:
		return object, nil
	}
}

// findPartialSetter finds the setter matching the name if one exists
func (fs *fieldSetter) findSetter(field *yaml.RNode) (
	*fieldmeta.PartialFieldSetter, *fieldmeta.FieldMeta, bool, error) {
	// check if there are any substitutions for this field
	var fm = &fieldmeta.FieldMeta{}
	if err := fm.Read(field); err != nil {
		return nil, nil, false, err
	}
	if fs.SetBy != "" {
		fm.Extensions.SetBy = fs.SetBy
	}
	if fs.Description != "" {
		fm.Schema.Description = fs.Description
	}

	if fm.Extensions.FieldSetter != nil && fm.Extensions.FieldSetter.Name == fs.Name {
		return fm.Extensions.FieldSetter, fm, false, nil
	}

	// check if there is a matching substitution
	for i := range fm.Extensions.PartialFieldSetters {
		if fm.Extensions.PartialFieldSetters[i].Name == fs.Name {
			return &fm.Extensions.PartialFieldSetters[i], fm, true, nil
		}
	}
	return nil, nil, false, nil
}

// set performs the substitution for the given field, substitution, and metadata
func (fs *fieldSetter) set(
	field *yaml.RNode, s *fieldmeta.PartialFieldSetter,
	f *fieldmeta.FieldMeta, partial bool) error {
	if s.Value == fs.Value || !strings.Contains(field.YNode().Value, s.Value) {
		// no substitutions necessary -- already substituted or doesn't have the set value
		// which acts as a marker
		return nil
	}

	// record that the config has been modified
	fs.Count++

	if !partial {
		// full setter
		field.YNode().Value = fs.Value
	} else {
		// replace the current value with the new value
		field.YNode().Value = strings.ReplaceAll(field.YNode().Value, s.Value, fs.Value)
	}

	// be sure to set the tag to the matching type so the yaml doesn't incorrectly quote
	//integers or booleans as strings
	fType := fieldmeta.FieldValueType(f.Schema.Type[0])
	if err := fType.Validate(field.YNode().Value); err != nil {
		return err
	}
	field.YNode().Tag = fType.Tag()

	// update the comment on the field
	s.Value = fs.Value
	if err := f.Write(field); err != nil {
		return err
	}
	return nil
}
