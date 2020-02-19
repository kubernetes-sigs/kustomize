// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Set sets resource field values from an OpenAPI setter
type Set struct {
	// Name is the name of the setter to set on the object.  i.e. matches the x-k8s-cli.setter.name
	// of the setter that should have its value applied to fields which reference it.
	Name string

	// Count is the number of fields that were updated by calling Filter
	Count int
}

// Filter implements Set as a yaml.Filter
func (s *Set) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	return object, accept(s, object)
}

// visitScalar
func (s *Set) visitScalar(object *yaml.RNode, p string) error {
	// get the openAPI for this field describing how to apply the setter
	ext, err := getExtFromComment(object)
	if err != nil {
		return err
	}
	if ext == nil {
		return nil
	}

	// perform a direct set of the field if it matches
	if s.set(object, ext) {
		s.Count++
		return nil
	}

	// perform a substitution of the field if it matches
	sub, err := s.substitute(object, ext)
	if err != nil {
		return err
	}
	if sub {
		s.Count++
	}
	return nil
}

// substitute updates the value of field from ext if ext contains a substitution that
// depends on a setter whose name matches s.Name.
func (s *Set) substitute(field *yaml.RNode, ext *cliExtension) (bool, error) {
	nameMatch := false

	// check partial setters to see if they contain the setter as part of a
	// substitution
	if ext.Substitution == nil {
		return false, nil
	}

	p := ext.Substitution.Pattern

	// substitute each setter into the pattern to get the new value
	for _, v := range ext.Substitution.Values {
		if v.Ref == "" {
			return false, errors.Errorf(
				"missing reference on substitution " + ext.Substitution.Name)
		}
		ref, err := spec.NewRef(v.Ref)
		if err != nil {
			return false, errors.Wrap(err)
		}
		setter, err := openapi.Resolve(&ref) // resolve the setter to its openAPI def
		if err != nil {
			return false, errors.Wrap(err)
		}
		subSetter, err := getExtFromSchema(setter) // parse the extension out of the openAPI
		if err != nil {
			return false, errors.Wrap(err)
		}
		// substitute the setters current value into the substitution pattern
		p = strings.ReplaceAll(p, v.Marker, subSetter.Setter.Value)

		if subSetter.Setter.Name == s.Name {
			// the substitution depends on the specified setter
			nameMatch = true
		}
	}
	if !nameMatch {
		// doesn't depend on the setter, don't modify its value
		return false, nil
	}

	// TODO(pwittrock): validate the field value

	field.YNode().Value = p
	return true, nil
}

// set applies the value from ext to field if its name matches s.Name
func (s *Set) set(field *yaml.RNode, ext *cliExtension) bool {
	// check full setter
	if ext.Setter == nil || ext.Setter.Name != s.Name {
		return false
	}

	// TODO(pwittrock): validate the field value

	// this has a full setter, set its value
	field.YNode().Value = ext.Setter.Value
	return true
}

// SetOpenAPI updates a setter value
type SetOpenAPI struct {
	// Name is the name of the setter to add
	Name string `yaml:"name"`
	// Value is the current value of the setter
	Value string `yaml:"value"`

	Description string `yaml:"description"`

	SetBy string `yaml:"setBy"`
}

// UpdateFile updates the OpenAPI definitions in a file with the given setter value.
func (s SetOpenAPI) UpdateFile(path string) error {
	return yaml.UpdateFile(s, path)
}

func (s SetOpenAPI) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	key := SetterDefinitionPrefix + s.Name
	def, err := object.Pipe(yaml.Lookup(
		"openAPI", "definitions", key, "x-k8s-cli", "setter"))
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, errors.Errorf("no setter %s found", s.Name)
	}
	if err := def.PipeE(&yaml.FieldSetter{Name: "value", StringValue: s.Value}); err != nil {
		return nil, err
	}

	if s.SetBy != "" {
		if err := def.PipeE(&yaml.FieldSetter{Name: "setBy", StringValue: s.SetBy}); err != nil {
			return nil, err
		}
	}

	if s.Description != "" {
		d, err := object.Pipe(yaml.LookupCreate(
			yaml.MappingNode, "openAPI", "definitions", key))
		if err != nil {
			return nil, err
		}
		if err := d.PipeE(&yaml.FieldSetter{Name: "description", StringValue: s.Description}); err != nil {
			return nil, err
		}
	}

	return object, nil
}
