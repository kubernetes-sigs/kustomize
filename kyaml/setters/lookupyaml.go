// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters

import (
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &lookupSetters{}

// lookupSetters looks up setters for a Resource
type lookupSetters struct {
	// Name of the setter to lookup.  Optional
	Name string

	// Setters is a list of setters that were found
	Setters []setter
}

type setter struct {
	fieldmeta.PartialFieldSetter
	Description string
	Type        string
	SetBy       string
}

func (ls *lookupSetters) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		// skip the document node
		return ls.Filter(yaml.NewRNode(object.YNode().Content[0]))
	case yaml.MappingNode:
		return object, object.VisitFields(func(node *yaml.MapNode) error {
			return node.Value.PipeE(ls)
		})
	case yaml.SequenceNode:
		return object, object.VisitElements(func(node *yaml.RNode) error {
			return node.PipeE(ls)
		})
	case yaml.ScalarNode:
		return object, ls.lookup(object)
	default:
		return object, nil
	}
}

// lookup finds any setters for a field
func (ls *lookupSetters) lookup(field *yaml.RNode) error {
	// check if there is a substitution for this field
	var fm = &fieldmeta.FieldMeta{}
	if err := fm.Read(field); err != nil {
		return err
	}

	if fm.Extensions.FieldSetter != nil {
		if ls.Name != "" && ls.Name != fm.Extensions.FieldSetter.Name {
			// skip this setter, it doesn't match the specified setter
			return nil
		}
		// full setter
		ls.Setters = append(ls.Setters, setter{
			PartialFieldSetter: *fm.Extensions.FieldSetter,
			Description:        fm.Schema.Description,
			Type:               fm.Schema.Type[0],
			SetBy:              fm.Extensions.SetBy,
		})
		return nil
	}

	// partial setters
	for i := range fm.Extensions.PartialFieldSetters {
		if ls.Name != "" && ls.Name != fm.Extensions.PartialFieldSetters[i].Name {
			// skip this setter
			continue
		}
		ls.Setters = append(ls.Setters, setter{
			PartialFieldSetter: fm.Extensions.PartialFieldSetters[i],
			Description:        fm.Schema.Description,
			Type:               fm.Schema.Type[0],
			SetBy:              fm.Extensions.SetBy,
		})
	}
	return nil
}
