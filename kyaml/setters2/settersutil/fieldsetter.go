// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"io/ioutil"
	"os"
	"path/filepath"

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

	OpenAPIFileName string

	ResourcesPath string

	RecurseSubPackages bool

	IsSet bool
}

func (fs *FieldSetter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	fs.Count, _ = fs.Set()
	return nil, nil
}

// Set updates the OpenAPI definitions and resources with the new setter value
func (fs FieldSetter) Set() (int, error) {
	// Update the OpenAPI definitions
	soa := setters2.SetOpenAPI{
		Name:        fs.Name,
		Value:       fs.Value,
		ListValues:  fs.ListValues,
		Description: fs.Description,
		SetBy:       fs.SetBy,
		IsSet:       fs.IsSet,
	}

	// the input field value is updated in the openAPI file and then parsed
	// at to get the value and set it to resource files, but if there is error
	// after updating openAPI file and while updating resources, the openAPI
	// file should be reverted, as set operation failed
	stat, err := os.Stat(fs.OpenAPIPath)
	if err != nil {
		return 0, err
	}

	curOpenAPI, err := ioutil.ReadFile(fs.OpenAPIPath)
	if err != nil {
		return 0, err
	}

	// write the new input value to openAPI file
	if err := soa.UpdateFile(fs.OpenAPIPath); err != nil {
		return 0, err
	}

	// Load the updated definitions
	clean, err := openapi.AddSchemaFromFile(fs.OpenAPIPath)
	if err != nil {
		return 0, err
	}
	defer clean()

	// Update the resources with the new value
	// Set NoDeleteFiles to true as SetAll will return only the nodes of files which should be updated and
	// hence, rest of the files should not be deleted
	inout := &kio.LocalPackageReadWriter{PackagePath: fs.ResourcesPath, NoDeleteFiles: true, PackageFileName: fs.OpenAPIFileName}
	s := &setters2.Set{Name: fs.Name}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{setters2.SetAll(s)},
		Outputs: []kio.Writer{inout},
	}.Execute()

	// revert openAPI file if set operation fails
	if err != nil {
		if writeErr := ioutil.WriteFile(fs.OpenAPIPath, curOpenAPI, stat.Mode().Perm()); writeErr != nil {
			return 0, writeErr
		}
	}
	return s.Count, err
}

// SetAllSetterDefinitions reads all the Setter Definitions from the input OpenAPI
// file and sets all values for the resource configs in the provided destination directories.
// If syncOpenAPI is true, the openAPI files in destination directories are also
// updated with the setter values in the input openAPI file
func SetAllSetterDefinitions(syncOpenAPI bool, openAPIPath string, dirs ...string) error {
	clean, err := openapi.AddSchemaFromFile(openAPIPath)
	if err != nil {
		return err
	}
	defer clean()
	for _, destDir := range dirs {
		if syncOpenAPI {
			openAPIFileName := filepath.Base(openAPIPath)
			err := syncOpenAPIValuesWithSchema(filepath.Join(destDir, openAPIFileName))
			if err != nil {
				return err
			}
		}
		rw := &kio.LocalPackageReadWriter{
			PackagePath: destDir,
			// set output won't include resources from files which
			// weren't modified.  make sure we don't delete them.
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

// syncOpenAPIValuesWithSchema syncs the setter values in global openAPI schema
// with the ones in openAPIPath and writes them back
func syncOpenAPIValuesWithSchema(openAPIPath string) error {
	refs, err := openapi.DefinitionRefs(openAPIPath)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		sch := openapi.Schema().Definitions[ref]
		cliExt, err := setters2.GetExtFromSchema(&sch)
		if cliExt == nil || cliExt.Setter == nil || err != nil {
			// if the ref doesn't exist in global schema or if it is not a setter
			// continue, as there might be setters which are not present global schema
			continue
		}
		soa := setters2.SetOpenAPI{
			Name:       cliExt.Setter.Name,
			Value:      cliExt.Setter.Value,
			ListValues: cliExt.Setter.ListValues,
			IsSet:      true,
		}
		if err := soa.UpdateFile(openAPIPath); err != nil {
			return err
		}
	}
	return nil
}
