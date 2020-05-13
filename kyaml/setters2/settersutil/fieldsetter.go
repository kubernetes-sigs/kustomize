// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"io/ioutil"
	"os"

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

	// the input field value is updated in the openAPI file and then parsed
	// at to get the value and set it to resource files, but if there is error
	// after updating openAPI file and while updating resources, the openAPI
	// file should be reverted, as set operation failed
	stat, err := os.Stat(openAPIPath)
	if err != nil {
		return 0, err
	}

	curOpenAPI, err := ioutil.ReadFile(openAPIPath)
	if err != nil {
		return 0, err
	}

	// write the new input value to openAPI file
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
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{setters2.SetAll(s)},
		Outputs: []kio.Writer{inout},
	}.Execute()

	// revert openAPI file if set operation fails
	if err != nil {
		if writeErr := ioutil.WriteFile(openAPIPath, curOpenAPI, stat.Mode().Perm()); writeErr != nil {
			return 0, writeErr
		}
	}
	return s.Count, err
}

// SetAllSetterDefinitions reads all the Setter Definitions from the OpenAPI
// file and sets all values in the provided directories.
func SetAllSetterDefinitions(openAPIPath string, dirs ...string) error {
	if err := openapi.AddSchemaFromFile(openAPIPath); err != nil {
		return err
	}

	for _, dir := range dirs {
		rw := &kio.LocalPackageReadWriter{
			PackagePath: dir,
			// set output won't include resources from files which
			//weren't modified.  make sure we don't delete them.
			NoDeleteFiles: true,
		}

		// apply all of the setters to the directory
		err := kio.Pipeline{
			Inputs: []kio.Reader{rw},
			// Set all of the setters
			Filters: []kio.Filter{setters2.SetAll(&setters2.Set{SetAll: true})},
			Outputs: []kio.Writer{rw},
		}.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}
