// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &Marker{}

// substituteResource substitutes a Marker value on a field
type Marker struct {
	// Path is the path of the field to add the substitution for
	Field string

	// Substitution is the substitution to add
	Substitution fieldmeta.Substitution

	// PartialMatch if true will match if the Substitution value is a substring of the current
	// value.
	PartialMatch bool

	Description string
	OwnedBy     string
	Type        string

	// currentFieldName is the name of the current field being processed
	currentFieldName string
}

// Filter performs the substitutions for a single object
func (m *Marker) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		return m.Filter(yaml.NewRNode(object.YNode().Content[0]))
	case yaml.MappingNode:
		return object, object.VisitFields(func(node *yaml.MapNode) error {
			// set the current field name
			n := m.currentFieldName
			defer func() { m.currentFieldName = n }()
			m.currentFieldName = node.Key.YNode().Value
			_, err := m.Filter(node.Value)
			return err
		})
	case yaml.SequenceNode:
		return object, object.VisitElements(func(node *yaml.RNode) error {
			_, err := m.Filter(node)
			return err
		})
	case yaml.ScalarNode:
		if m.currentFieldName != m.Field {
			return object, nil
		}
		if err := m.createSub(object); err != nil {
			return nil, err
		}
		return object, nil
	default:
		return object, nil
	}
}

func (m *Marker) createSub(field *yaml.RNode) error {
	// doesn't match the supplied value
	if field.YNode().Value != m.Substitution.Value {
		if !m.PartialMatch || !strings.Contains(field.YNode().Value, m.Substitution.Value) {
			return nil
		}
	}

	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(field); err != nil {
		return errors.Wrap(err)
	}
	fm.OwnedBy = m.OwnedBy
	fm.Description = m.Description
	fm.Type = fieldmeta.FieldValueType(m.Type)
	if m.Substitution.Marker == "" {
		m.Substitution.Marker = "[MARKER]"
	}

	found := false
	for i := range fm.Substitutions {
		s := fm.Substitutions[i]
		if s.Name == m.Substitution.Name {
			// update the substitution if we find it
			found = true
			fm.Substitutions[i] = m.Substitution
			break
		}
	}
	if !found {
		// add the substitution if it wasn't found
		fm.Substitutions = append(fm.Substitutions, m.Substitution)
	}
	if err := fm.Write(field); err != nil {
		return errors.Wrap(err)
	}
	return nil
}
