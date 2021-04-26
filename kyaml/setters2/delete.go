// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Delete delete openAPI definition references from resource fields.
type Delete struct {
	// Name is the name of the openAPI definition to delete.
	Name string

	// DefinitionPrefix is the prefix of the OpenAPI definition type
	DefinitionPrefix string

	SettersSchema *spec.Schema
}

// Filter implements yaml.Filter
func (d *Delete) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	return object, accept(d, object, d.SettersSchema)
}

func (d *Delete) visitSequence(_ *yaml.RNode, _ string, _ *openapi.ResourceSchema) error {
	// no-op
	return nil
}

func (d *Delete) visitMapping(object *yaml.RNode, _ string, _ *openapi.ResourceSchema) error {
	fieldRNodes, err := object.FieldRNodes()
	if err != nil {
		return err
	}

	// for each of the field node key visit it as scalar to delete the array setter comment
	for _, fieldRNode := range fieldRNodes {
		err := d.visitScalar(fieldRNode, "", nil, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// visitScalar implements visitor
// visitScalar will remove the reference on each scalar field whose name matches.
func (d *Delete) visitScalar(object *yaml.RNode, _ string, _, _ *openapi.ResourceSchema) error {
	// read the field metadata
	fm := fieldmeta.FieldMeta{SettersSchema: d.SettersSchema}
	if err := fm.Read(object); err != nil {
		return err
	}

	// Delete the reference iff the ref string matches with DefinitionPrefix
	if strings.HasSuffix(fm.Schema.Ref.String(), d.DefinitionPrefix+d.Name) {
		// remove the ref on the metadata
		fm.Schema.Ref = spec.Ref{}

		// write the field metadata
		if err := fm.Write(object); err != nil {
			return err
		}
	}

	return nil
}

// DeleterDefinition may be used to update a files OpenAPI definitions with a new setter.
type DeleterDefinition struct {
	// Name is the name of the openAPI definition to delete.
	Name string `yaml:"name"`

	// DefinitionPrefix is the prefix of the OpenAPI definition type
	DefinitionPrefix string `yaml:"definitionPrefix"`
}

func (dd DeleterDefinition) DeleteFromFile(path string) error {
	return yaml.UpdateFile(dd, path)
}

// SubstReferringDefinition check if the definition used in substitution and return the substitution name if true
func SubstReferringDefinition(definitions *yaml.RNode, key string) string {
	fieldNames, err := definitions.Fields()
	if err != nil {
		return ""
	}
	for _, fieldName := range fieldNames {
		// the definition key -- contains the substitution name
		subkey := definitions.Field(fieldName).Key.YNode().Value
		if strings.HasPrefix(subkey, fieldmeta.SubstitutionDefinitionPrefix) {
			substNode, err := definitions.Field(fieldName).Value.Pipe(yaml.Lookup(K8sCliExtensionKey, "substitution"))
			if err != nil {
				continue
			}

			b, err := substNode.MarshalJSON()
			if err != nil {
				continue
			}

			subst := SubstitutionDefinition{}
			if err := yaml.Unmarshal(b, &subst); err != nil {
				continue
			}
			// Check the ref in value to see if it contains the setter key
			for _, v := range subst.Values {
				if strings.HasSuffix(v.Ref, key) {
					return subst.Name
				}
			}
		}
	}

	return ""
}

func (dd DeleterDefinition) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	key := dd.DefinitionPrefix + dd.Name
	var defType string

	switch dd.DefinitionPrefix {
	case fieldmeta.SubstitutionDefinitionPrefix:
		defType = "substitution"
	case fieldmeta.SetterDefinitionPrefix:
		defType = "setter"
	default:
		return nil, errors.Errorf("the input delete definitionPrefix does't match any of openAPI definitions, "+
			"allowed values [%s, %s]", fieldmeta.SetterDefinitionPrefix, fieldmeta.SubstitutionDefinitionPrefix)
	}

	definitions, err := object.Pipe(yaml.Lookup(openapi.SupplementaryOpenAPIFieldName, "definitions"))
	if err != nil {
		return nil, err
	}
	// return error if the setter to be deleted doesn't exist
	if definitions == nil || definitions.Field(key) == nil {
		return nil, errors.Errorf("%s %q does not exist", defType, dd.Name)
	}

	subst := SubstReferringDefinition(definitions, key)

	if subst != "" {
		return nil, errors.Errorf("%s %q is used in substitution %q, please delete the parent substitution first", defType, dd.Name, subst)
	}

	_, err = definitions.Pipe(yaml.FieldClearer{Name: key})
	if err != nil {
		return nil, err
	}
	// remove definitions if it's empty
	_, err = object.Pipe(yaml.Lookup(openapi.SupplementaryOpenAPIFieldName), yaml.FieldClearer{Name: "definitions", IfEmpty: true})
	if err != nil {
		return nil, err
	}

	// remove openApi if it's empty
	_, err = object.Pipe(yaml.FieldClearer{Name: openapi.SupplementaryOpenAPIFieldName, IfEmpty: true})
	if err != nil {
		return nil, err
	}

	return object, nil
}
