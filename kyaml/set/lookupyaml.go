// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &lookupSubstitutions{}

// substituteResource substitutes a Marker value on a field
type lookupSubstitutions struct {
	// Name of the substitution to lookup.  If unspecified lookup all substitutions.
	Name string

	// FieldSubstitution is the list of substitutions that were found
	Substitutions []FieldSubstitution
}

func (ls *lookupSubstitutions) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		return ls.Filter(yaml.NewRNode(object.YNode().Content[0]))
	case yaml.MappingNode:
		return object, object.VisitFields(func(node *yaml.MapNode) error {
			_, err := ls.Filter(node.Value)
			return err
		})
	case yaml.SequenceNode:
		return object, object.VisitElements(func(node *yaml.RNode) error {
			_, err := ls.Filter(node)
			return err
		})
	case yaml.ScalarNode:
		return object, ls.lookup(object)
	default:
		return object, nil
	}
}

// lookup finds any substitutions for this field
func (ls *lookupSubstitutions) lookup(field *yaml.RNode) error {
	// check if there is a substitution for this field
	var fm = &fieldmeta.FieldMeta{}
	if err := fm.Read(field); err != nil {
		return err
	}

	for i := range fm.Substitutions {
		s := fm.Substitutions[i]
		if ls.Name == "" || ls.Name == s.Name {
			ls.Substitutions = append(ls.Substitutions, FieldSubstitution{
				Name:         s.Name,
				CurrentValue: s.Value,
				Description:  fm.Description,
				Marker:       s.Marker,
				Type:         fm.Type,
				OwnedBy:      fm.OwnedBy,
			})
		}
	}
	return nil
}
