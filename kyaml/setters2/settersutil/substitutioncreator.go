// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2"
)

// SubstitutionCreator creates or updates a substitution in the OpenAPI definitions, and
// inserts references to the substitution from matching resource fields.
type SubstitutionCreator struct {
	// Name is the name of the substitution to create
	Name string

	// Pattern is the substitution pattern
	Pattern string

	// Values are the substitution values for the pattern
	Values []setters2.Value

	// FieldName if set will add the OpenAPI reference to fields with this name or path
	// FieldName may be the full name of the field, full path to the field, or the path suffix.
	// e.g. all of the following would match spec.template.spec.containers.image --
	// [image, containers.image, spec.containers.image, template.spec.containers.image,
	//  spec.template.spec.containers.image]
	// Optional.  If unspecified match all field names.
	FieldName string

	// FieldValue if set will add the OpenAPI reference to fields if they have this value.
	// Optional.  If unspecified match all field values.
	FieldValue string
}

func (c SubstitutionCreator) Create(openAPIPath, resourcesPath string) error {
	d := setters2.SubstitutionDefinition{
		Name:    c.Name,
		Values:  c.Values,
		Pattern: c.Pattern,
	}
	if err := d.AddToFile(openAPIPath); err != nil {
		return err
	}

	// Load the updated definitions
	if err := openapi.AddSchemaFromFile(openAPIPath); err != nil {
		return err
	}

	// Update the resources with the setter reference
	inout := &kio.LocalPackageReadWriter{PackagePath: resourcesPath}
	return kio.Pipeline{
		Inputs: []kio.Reader{inout},
		Filters: []kio.Filter{kio.FilterAll(
			&setters2.Add{
				FieldName:  c.FieldName,
				FieldValue: c.FieldValue,
				Ref:        setters2.DefinitionsPrefix + setters2.SubstitutionDefinitionPrefix + c.Name,
			})},
		Outputs: []kio.Writer{inout},
	}.Execute()
}
