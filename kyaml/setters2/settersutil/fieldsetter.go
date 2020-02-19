// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2"
)

// FieldSetter sets the value for a field setter.
type FieldSetter struct {
	// Name is the name of the setter to set
	Name string

	// Value is the value to set
	Value string

	Description string

	SetBy string
}

// Set updates the OpenAPI definitions and resources with the new setter value
func (fs FieldSetter) Set(openAPIPath, resourcesPath string) (int, error) {
	// Update the OpenAPI definitions
	soa := setters2.SetOpenAPI{
		Name: fs.Name, Value: fs.Value, Description: fs.Description, SetBy: fs.SetBy}
	if err := soa.UpdateFile(openAPIPath); err != nil {
		return 0, err
	}

	// Load the updated definitions
	if err := openapi.AddSchemaFromFile(openAPIPath); err != nil {
		return 0, err
	}

	// Update the resources with the new value
	inout := &kio.LocalPackageReadWriter{PackagePath: resourcesPath}
	s := &setters2.Set{Name: fs.Name}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{kio.FilterAll(s)},
		Outputs: []kio.Writer{inout},
	}.Execute()
	return s.Count, err
}
