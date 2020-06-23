// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Delete delete setter or substitution references from resource fields.
// Requires that FieldName have been set.
type Delete struct {

	// FieldName if delete the OpenAPI reference to fields with this name or path
	// FieldName may be the full name of the field, full path to the field, or the path suffix.
	// e.g. all of the following would match spec.template.spec.containers.image --
	// [image, containers.image, spec.containers.image, template.spec.containers.image,
	//  spec.template.spec.containers.image]
	FieldName string
}

// Filter implements yaml.Filter
func (d *Delete) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	if d.FieldName == "" {
		return nil, errors.Errorf("must specify fieldName")
	}
	return object, accept(d, object)
}

func (d *Delete) visitSequence(_ *yaml.RNode, _ string, _ *openapi.ResourceSchema) error {
	// no-op
	return nil
}

func (d *Delete) visitMapping(_ *yaml.RNode, _ string, _ *openapi.ResourceSchema) error {
	// no-op
	return nil
}

// visitScalar implements visitor
// visitScalar will remove the reference on each scalar field whose name matches.
func (d *Delete) visitScalar(object *yaml.RNode, p string, _ *openapi.ResourceSchema) error {
	// check if the field matches
	if d.FieldName != "" && !strings.HasSuffix(p, d.FieldName) {
		return nil
	}

	// read the field metadata
	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(object); err != nil {
		return err
	}

	// remove the ref on the metadata
	fm.Schema.Ref = spec.Ref{}

	// write the field metadata
	if err := fm.Write(object); err != nil {
		return err
	}

	return nil
}

// DeleterDefinition may be used to update a files OpenAPI definitions with a new setter.
type DeleterDefinition struct {
	// Name is the name of the setter to create or update.
	Name string `yaml:"name"`
}

func (dd DeleterDefinition) DeleteFromFile(path string) error {
	return yaml.UpdateFile(dd, path)
}

// SubstReferringSetter check if the setter used in substitution and return the substitution name if true
func SubstReferringSetter(definitions *yaml.RNode, key string) string {
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
	key := fieldmeta.SetterDefinitionPrefix + dd.Name

	definitions, err := object.Pipe(yaml.Lookup(openapi.SupplementaryOpenAPIFieldName, "definitions"))
	if err != nil || definitions == nil {
		return nil, err
	}
	// return error if the setter to be deleted doesn't exist
	if definitions.Field(key) == nil {
		return nil, errors.Errorf("setter does not exist")
	}

	subst := SubstReferringSetter(definitions, key)

	if subst != "" {
		return nil, errors.Errorf("setter is used in substitution %s, please delete the substitution first", subst)
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
