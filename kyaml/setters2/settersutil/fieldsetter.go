// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FieldSetter sets the value for a field setter.
type FieldSetter struct {
	// Name is the name of the setter to set
	Name string

	// Value is the value to set
	Value string

	// ListValues contains a list of values to set on a Sequence
	ListValues []string

	Description string

	SetBy string

	Count int

	OpenAPIPath string

	ResourcesPath string
}

func (fs *FieldSetter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	fs.Count, _ = fs.Set(fs.OpenAPIPath, fs.ResourcesPath)
	return nil, nil
}

// Set updates the OpenAPI definitions and resources with the new setter value
func (fs FieldSetter) Set(openAPIPath, resourcesPath string) (int, error) {
	// Update the OpenAPI definitions
	soa := setters2.SetOpenAPI{
		Name:        fs.Name,
		Value:       fs.Value,
		ListValues:  fs.ListValues,
		Description: fs.Description,
		SetBy:       fs.SetBy,
	}
	if err := soa.UpdateFile(openAPIPath); err != nil {
		return 0, err
	}

	// Load the updated definitions
	if err := openapi.AddSchemaFromFile(openAPIPath); err != nil {
		return 0, err
	}

	// Update the resources with the new value
	// Set NoDeleteFiles to true as SetAll will return only the nodes of files which should be updated and
	// hence, rest of the files should not be deleted
	inout := &kio.LocalPackageReadWriter{PackagePath: resourcesPath, NoDeleteFiles: true}
	s := &setters2.Set{Name: fs.Name}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{setters2.SetAll(s)},
		Outputs: []kio.Writer{inout},
	}.Execute()
	return s.Count, err
}
