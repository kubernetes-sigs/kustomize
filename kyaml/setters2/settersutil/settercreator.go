// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2"
)

// SetterCreator creates or updates a setter in the OpenAPI definitions, and inserts references
// to the setter from matching resource fields.
type SetterCreator struct {
	// Name is the name of the setter to create or update.
	Name string

	Description string

	SetBy string

	Type string

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

func (c SetterCreator) Create(openAPIPath, resourcesPath string) error {
	// Update the OpenAPI definitions to hace the setter
	sd := setters2.SetterDefinition{
		Name: c.Name, Value: c.FieldValue, Description: c.Description, SetBy: c.SetBy,
		Type: c.Type,
	}
	if err := sd.AddToFile(openAPIPath); err != nil {
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
				Ref:        setters2.DefinitionsPrefix + setters2.SetterDefinitionPrefix + c.Name,
			})},
		Outputs: []kio.Writer{inout},
	}.Execute()
}
