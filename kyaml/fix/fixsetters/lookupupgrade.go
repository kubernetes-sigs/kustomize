// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fixsetters

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ yaml.Filter = &upgradeV1Setters{}

// upgradeV1Setters looks up v1 setters in a Resource and upgrades
// all the setters related comments
type upgradeV1Setters struct {
	// Name of the setter to lookup.  Optional
	Name string

	// Setters is a list of setters that were found
	Setters []setter

	// Substitutions is a list of substitutions that were found
	Substitutions []substitution
}

type substitution struct {
	Name      string
	FieldVale string
	Pattern   string
}

type setter struct {
	PartialFieldSetter
	Description string
	Type        string
	SetBy       string
	SubstName   string
}

func (ls *upgradeV1Setters) Filter(object *yaml.RNode) (*yaml.RNode, error) {
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
		return object, ls.lookupAndUpgrade(object)
	default:
		return object, nil
	}
}

// lookupAndUpgrade finds any setters for a field and upgrades the setters comment
func (ls *upgradeV1Setters) lookupAndUpgrade(field *yaml.RNode) error {
	// check if there is a substitution for this field
	var fm = FieldMetaV1{}
	if err := fm.UpgradeV1SetterComment(field); err != nil {
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

	if len(fm.Extensions.PartialFieldSetters) > 0 {
		substName := "subst"
		filedValue := field.YNode().Value
		pattern := filedValue

		// derive substitution pattern from partial setters
		for i := range fm.Extensions.PartialFieldSetters {
			substName += "-" + fm.Extensions.PartialFieldSetters[i].Name
			pattern = strings.Replace(pattern, fm.Extensions.PartialFieldSetters[i].Value, `${`+fm.Extensions.PartialFieldSetters[i].Name+"}", 1)
		}

		ls.Substitutions = append(ls.Substitutions, substitution{
			Name:      substName,
			FieldVale: filedValue,
			Pattern:   pattern,
		})
	}

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
