// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &customFieldSetter{}

// customFieldSetter creates a new custom field setter
type customFieldSetter struct {
	// Path is the path of the field to add the setter for
	Field string

	// Setter is the setter to add
	Setter fieldmeta.PartialFieldSetter

	// Description is the description to add to the OpenAPI
	Description string

	// SetBy is the setBy to add to the OpenAPI extension
	SetBy string

	Type string

	// currentFieldName is the name of the current field being processed
	currentFieldName string
}

// Filter performs the setter for a single object
func (m *customFieldSetter) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		return m.Filter(yaml.NewRNode(object.YNode().Content[0]))
	case yaml.MappingNode:
		return object, object.VisitFields(func(node *yaml.MapNode) error {
			// record the current field name, resetting it back to its original value
			// when done
			n := m.currentFieldName
			defer func() { m.currentFieldName = n }()
			m.currentFieldName = node.Key.YNode().Value
			return node.Value.PipeE(m)
		})
	case yaml.SequenceNode:
		return object, object.VisitElements(func(node *yaml.RNode) error {
			return node.PipeE(m)
		})
	case yaml.ScalarNode:
		// only create the setter for fields with the given name
		if m.currentFieldName != m.Field {
			return object, nil
		}
		if err := m.create(object); err != nil {
			return nil, err
		}
		return object, nil
	default:
		return object, nil
	}
}

func (m *customFieldSetter) create(field *yaml.RNode) error {
	// doesn't match the supplied value
	if !strings.Contains(field.YNode().Value, m.Setter.Value) {
		return nil
	}

	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(field); err != nil {
		return errors.Wrap(err)
	}

	if m.Description != "" {
		fm.Schema.Description = m.Description
	}

	fm.Extensions.SetBy = m.SetBy
	fm.Schema.Type = []string{m.Type}

	found := false
	for i := range fm.Extensions.PartialFieldSetters {
		s := fm.Extensions.PartialFieldSetters[i]
		if s.Name == m.Setter.Name {
			// update the setter if we find it
			found = true
			fm.Extensions.PartialFieldSetters[i] = m.Setter
			break
		}
	}
	if !found {
		// add the setter if it wasn't found
		fm.Extensions.PartialFieldSetters = append(fm.Extensions.PartialFieldSetters, m.Setter)
	}
	if err := fm.Write(field); err != nil {
		return errors.Wrap(err)
	}
	return nil
}
