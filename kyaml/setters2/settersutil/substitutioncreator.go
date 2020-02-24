// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"io/ioutil"
	"math"
	"strings"

	"sigs.k8s.io/kind/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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

	err := c.CreateSettersForSubstitution(openAPIPath)
	if err != nil {
		return err
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

// CreateSettersForSubstitution creates the setters for all the references in the substitution
// values if they don't already exist in openAPIPath file.
func (c SubstitutionCreator) CreateSettersForSubstitution(openAPIPath string) error {
	b, err := ioutil.ReadFile(openAPIPath)
	if err != nil {
		return err
	}

	// parse the yaml file
	y, err := yaml.Parse(string(b))
	if err != nil {
		return err
	}

	m, err := c.GetValuesForMarkers()
	if err != nil {
		return err
	}

	// for each ref in values, check if the setter already exists, if not create them
	for _, value := range c.Values {
		obj, err := y.Pipe(yaml.Lookup(
			// get the setter key from ref. Ex: from #/definitions/io.k8s.cli.setters.image_setter
			// extract io.k8s.cli.setters.image_setter
			"openAPI", "definitions", strings.TrimPrefix(value.Ref, "#/definitions/")))

		if err != nil {
			return err
		}

		if obj == nil {
			sd := setters2.SetterDefinition{
				// get the setter name from ref. Ex: from #/definitions/io.k8s.cli.setters.image_setter
				// extract image_setter
				Name:  strings.TrimPrefix(value.Ref, "#/definitions/io.k8s.cli.setters."),
				Value: m[value.Marker],
			}
			err := sd.AddToFile(openAPIPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetValuesForMarkers parses the pattern and field value to derive values for the
// markers in the pattern string. Returns error if the marker values can't be derived
func (c SubstitutionCreator) GetValuesForMarkers() (map[string]string, error) {
	m := make(map[string]string)
	indices, err := c.GetStartIndices()
	if err != nil {
		return nil, err
	}
	s := c.FieldValue
	p := c.Pattern
	i := 0
	j := 0
	// iterate s, p with indices i, j respectively and when j hits the index of a marker, freeze j and iterate
	// i and capture string till we find the substring just after current marker and before next marker

	// Ex: s = "something/ubuntu:0.1.0", p = "something/IMAGE::VERSION", till j reaches 10
	// just proceed i and j and check if s[i]==p[j]
	// when j is 10, freeze j and move i till it sees substring '::' which derives IMAGE = ubuntu and so on.
	for i < len(s) && j < len(p) {
		if marker, ok := indices[j]; ok {
			value := ""
			e := j + len(marker)

			for i < len(s) && (e == len(p) ||
				s[i:min(len(s), i+lenToNextMarker(indices, e))] != p[e:min(e+lenToNextMarker(indices, e), len(p))]) {
				value += string(s[i])
				i++
			}
			// if marker is repeated in the pattern, make sure that the corresponding values
			// are same and throw error if not.
			if prevValue, ok := m[marker]; ok && prevValue != value {
				return nil, errors.Errorf("Same marker is found to have different values in field value.")
			}
			m[marker] = value
			j += len(marker)
		} else {
			if s[i] != p[j] {
				return nil, errors.Errorf("Unable to derive values for markers. Create setters for all markers and then try again.")
			}
			i++
			j++
		}
	}
	// check if both strings are completely visited or throw error
	if i < len(s) || j < len(p) {
		return nil, errors.Errorf("Unable to derive values for markers. Create setters for all markers and then try again.")
	}
	return m, nil
}

// GetStartIndices returns the start indices of all the markers in the pattern
func (c SubstitutionCreator) GetStartIndices() (map[int]string, error) {
	inds := make(map[int]string)
	p := c.Pattern
	for _, value := range c.Values {
		m := value.Marker
		found := false
		for i := range p {
			if strings.HasPrefix(p[i:], m) {
				inds[i] = m
				found = true
			}
		}
		if !found {
			return nil, errors.Errorf("Unable to find marker " + m + " in the pattern")
		}
	}
	return inds, nil
}

// lenToNextMarker takes in the indices map, an index and returns the distance to
// next greater index
func lenToNextMarker(m map[int]string, j int) int {
	res := math.MaxInt32
	for k := range m {
		if k > j {
			res = min(k-j, res)
		}
	}
	return res
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
