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
}

// Filter implements Set as a yaml.Filter
func (s *Set) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	return object, accept(s, object)
}

// visitScalar
func (s *Set) visitScalar(object *yaml.RNode, _ string) error {
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
		return nil
	}

	// perform a substitution of the field if it matches
	if sub, err := s.substitute(object, ext); sub || err != nil {
		return err
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
