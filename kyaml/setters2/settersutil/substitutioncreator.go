// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
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
	y, err := yaml.ReadFile(openAPIPath)
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
	fv := c.FieldValue
	pattern := c.Pattern
	fvInd := 0
	patternInd := 0
	// iterate fv, pattern with indices fvInd, patternInd respectively and when patternInd hits the index of a marker,
	// freeze patternInd and iterate fvInd and capture string till we find the substring just after current marker
	// and before next marker

	// Ex: fv = "something/ubuntu:0.1.0", pattern = "something/IMAGE:VERSION", till patternInd reaches 10
	// just proceed fvInd and patternInd and check if fv[fvInd]==pattern[patternInd] when patternInd is 10,
	// freeze patternInd and move fvInd till it sees substring ':' which derives IMAGE = ubuntu and so on.
	for fvInd < len(fv) && patternInd < len(pattern) {
		// if we hit marker index, extract its corresponding value
		if marker, ok := indices[patternInd]; ok {
			// increment the patternInd to end of marker. This helps us to extract the substring before next marker.
			patternInd += len(marker)
			var value string
			if value, fvInd, err = c.extractValueForMarker(fvInd, fv, patternInd, indices); err != nil {
				return nil, err
			}
			// if marker is repeated in the pattern, make sure that the corresponding values
			// are same and throw error if not.
			if prevValue, ok := m[marker]; ok && prevValue != value {
				return nil, errors.Errorf(
					"marker %s is found to have different values %s and %s", marker, prevValue, value)
			}
			m[marker] = value
		} else {
			// Ex: fv = "samething/ubuntu:0.1.0" pattern = "something/IMAGE:VERSION". Error out at 'a' in fv.
			if fv[fvInd] != pattern[patternInd] {
				return nil, errors.Errorf(
					"unable to derive values for markers, " +
						"create setters for all markers and then try again")
			}
			fvInd++
			patternInd++
		}
	}
	// check if both strings are completely visited or throw error
	if fvInd < len(fv) || patternInd < len(pattern) {
		return nil, errors.Errorf(
			"unable to derive values for markers, " +
				"create setters for all markers and then try again")
	}
	return m, nil
}

// GetStartIndices returns the start indices of all the markers in the pattern
func (c SubstitutionCreator) GetStartIndices() (map[int]string, error) {
	indices := make(map[int]string)
	for _, value := range c.Values {
		found := false
		for i := range c.Pattern {
			if strings.HasPrefix(c.Pattern[i:], value.Marker) {
				indices[i] = value.Marker
				found = true
			}
		}
		if !found {
			return nil, errors.Errorf("unable to find marker " + value.Marker + " in the pattern")
		}
	}
	if err := validateMarkers(indices); err != nil {
		return nil, err
	}
	return indices, nil
}

// validateMarkers takes the indices map, checks if any of 2 markers not have delimiters,
// checks if any marker is substring of other and returns error
func validateMarkers(indices map[int]string) error {
	for k1, v1 := range indices {
		for k2, v2 := range indices {
			if k1 != k2 && k1+len(v1) == k2 {
				return errors.Errorf(
					"markers %s and %s are found to have no delimiters between them,"+
						" pre-create setters and try again", v1, v2)
			}
			if v1 != v2 && strings.Contains(v1, v2) {
				return errors.Errorf(
					"markers %s is substring of %s,"+
						" no marker should be substring of other", v2, v1)
			}
		}
	}
	return nil
}

// extractValueForMarker returns the value string for a marker and the incremented index
func (c SubstitutionCreator) extractValueForMarker(fvInd int, fv string, patternInd int, indices map[int]string) (string, int, error) {
	nonMarkerStr := strTillNextMarker(indices, patternInd, c.Pattern)

	// return the remaining string of fv till end if patternInd is at end of pattern
	if patternInd == len(c.Pattern) {
		return fv[fvInd:], len(fv), nil
	}

	// split remaining fv starting from fvInd with the non marker substring delimiter and get the first value
	// In example fv = "something/ubuntu::0.1.0", pattern = "something/IMAGE::VERSION",
	// split with "::" delimiter in fv which gives markerValue = ubuntu for marker IMAGE
	// increment fvInd by length of extracted marker value and return fvInd
	if markerValues := strings.Split(fv[fvInd:], nonMarkerStr); len(markerValues) > 0 {
		return markerValues[0], fvInd + len(markerValues[0]), nil
	}

	return "", -1, errors.Errorf(
		"unable to derive values for markers," +
			" create setters for all markers and then try again")
}

// substrOfLen takes a string, start index and length and returns substring of given length
// or till end of string
func substrOfLen(str string, startInd int, length int) string {
	return str[startInd:min(len(str), startInd+length)]
}

// strTillNextMarker takes in the indices map, a start index and returns the substring till
// start of next marker
func strTillNextMarker(indices map[int]string, startInd int, pattern string) string {
	// initialize with max value which is length of pattern
	nextMarkerStartInd := len(pattern)
	for ind := range indices {
		if ind > startInd {
			nextMarkerStartInd = min(ind-startInd, nextMarkerStartInd)
		}
	}
	return substrOfLen(pattern, startInd, nextMarkerStartInd)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
