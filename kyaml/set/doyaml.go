// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sub substitutes strings in fields
package set

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &performSubstitutions{}

// substituteResource substitutes a Marker value on a field
type performSubstitutions struct {
	// Name of the substitution to perform.
	Name string

	// Override if set to true will replace previously substituted values
	Override bool

	// Revert if set to true will undo previously substituted values
	Revert bool

	// NewValue is the new value to set.  Mutually exclusive with Revert.
	NewValue string

	// Description, if set will annotate the field with a description.
	Description string

	// OwnedBy, if set will annotate the field with an owner.
	OwnedBy string

	// Count will be incremented for each substituted value.
	Count int
}

// Filter performs the substitutions for a single object
func (fs *performSubstitutions) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		return fs.Filter(yaml.NewRNode(object.YNode().Content[0]))
	case yaml.MappingNode:
		return object, object.VisitFields(func(node *yaml.MapNode) error {
			_, err := fs.Filter(node.Value)
			return err
		})
	case yaml.SequenceNode:
		return object, object.VisitElements(func(node *yaml.RNode) error {
			_, err := fs.Filter(node)
			return err
		})
	case yaml.ScalarNode:
		s, f, err := fs.findSub(object)
		if err != nil {
			return nil, err
		}
		if s == nil {
			return object, nil
		}
		return object, fs.substitute(object, s, f)
	default:
		return object, nil
	}
}

// findSub finds the substitution matching the name if one exists
func (fs *performSubstitutions) findSub(field *yaml.RNode) (
	*fieldmeta.Substitution, *fieldmeta.FieldMeta, error) {
	// check if there are any substitutions for this field
	var fm = &fieldmeta.FieldMeta{}
	if err := fm.Read(field); err != nil {
		return nil, nil, err
	}
	if fs.OwnedBy != "" {
		fm.OwnedBy = fs.OwnedBy
	}
	if fs.Description != "" {
		fm.Description = fs.Description
	}

	// check if there is a matching substitution
	for i := range fm.Substitutions {
		if fm.Substitutions[i].Name == fs.Name {
			// validate the value if we are not reverting to the marker.
			// markers are allowed to be invalid.
			// only validate if there is a substitution matching the name
			if !fs.Revert {
				if err := fm.Type.Validate(fs.NewValue); err != nil {
					return nil, nil, err
				}
			}
			return &fm.Substitutions[i], fm, nil
		}
	}
	return nil, nil, nil
}

// substitute performs the substitution for the given field, substitution, and metadata
func (fs *performSubstitutions) substitute(
	field *yaml.RNode, s *fieldmeta.Substitution, f *fieldmeta.FieldMeta) error {
	// undo or override previous substitutions by substituting the marker back
	// NOTE: check if s.Value != "" so we never try to substitute the empty string back
	if (fs.Revert || fs.Override) && s.Value != "" {
		// revert to the marker value
		if strings.Contains(field.YNode().Value, s.Value) {
			// revert the substitution
			field.YNode().Value = strings.ReplaceAll(field.YNode().Value, s.Value, s.Marker)
			// only use the tag matching the type if the marker parses to that type
			field.YNode().Tag = f.Type.TagForValue(s.Marker)
			// record that the config has been modified
		}
	}
	if fs.Revert {
		fs.Count++
		s.Value = "" // value has been cleared and replaced with marker
		if err := f.Write(field); err != nil {
			return err
		}
		return nil
	}

	if s.Value == fs.NewValue || !strings.Contains(field.YNode().Value, s.Marker) {
		// no substitutions necessary -- already substituted or doesn't have the marker
		return nil
	}

	// replace the marker with the new value
	field.YNode().Value = strings.ReplaceAll(field.YNode().Value, s.Marker, fs.NewValue)
	// be sure to set the tag so the yaml doesn't incorrectly quote ints, bools or floats
	field.YNode().Tag = f.Type.Tag()
	field.YNode().Style = 0
	// record that the config has been modified
	fs.Count++

	// update the comment on the field
	s.Value = fs.NewValue
	if err := f.Write(field); err != nil {
		return err
	}
	return nil
}
