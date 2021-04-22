// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/sets"
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

	// Path to openAPI file
	OpenAPIPath string

	OpenAPIFileName string

	RecurseSubPackages bool

	// Path to resources folder
	ResourcesPath string

	SettersSchema *spec.Schema
}

func (c *SubstitutionCreator) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	return nil, c.Create()
}

func (c SubstitutionCreator) Create() error {
	err := c.validateSubstitutionInfo()
	if err != nil {
		return err
	}
	values, err := c.markersAndRefs(c.Name, c.Pattern)
	if err != nil {
		return err
	}
	c.Values = values
	d := setters2.SubstitutionDefinition{
		Name:    c.Name,
		Values:  c.Values,
		Pattern: c.Pattern,
	}

	// the input substitution definition is updated in the openAPI file and then parsed
	// to check if there are any cycles in nested substitutions, if there are
	// any, the openAPI file will be reverted to current state and error is thrown
	stat, err := os.Stat(c.OpenAPIPath)
	if err != nil {
		return err
	}

	curOpenAPI, err := ioutil.ReadFile(c.OpenAPIPath)
	if err != nil {
		return err
	}

	if err := d.AddToFile(c.OpenAPIPath); err != nil {
		return err
	}

	// Load the updated definitions
	sc, err := openapi.SchemaFromFile(c.OpenAPIPath)
	if err != nil {
		return err
	}
	c.SettersSchema = sc

	visited := sets.String{}
	ref, err := spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + c.Name)
	if err != nil {
		return err
	}

	schema, err := openapi.Resolve(&ref, c.SettersSchema)
	if err != nil {
		return err
	}

	ext, err := setters2.GetExtFromSchema(schema)
	if err != nil {
		return err
	}

	err = c.CreateSettersForSubstitution(c.OpenAPIPath)
	if err != nil {
		return err
	}

	// Load the updated definitions after setters are created
	sc, err = openapi.SchemaFromFile(c.OpenAPIPath)
	if err != nil {
		return err
	}
	c.SettersSchema = sc

	// revert openAPI file if there are cycles detected in created input substitution
	if err := c.checkForCycles(ext, visited); err != nil {
		if writeErr := ioutil.WriteFile(c.OpenAPIPath, curOpenAPI, stat.Mode().Perm()); writeErr != nil {
			return writeErr
		}
		return err
	}

	a := &setters2.Add{
		FieldName:     c.FieldName,
		FieldValue:    c.FieldValue,
		Ref:           fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + c.Name,
		SettersSchema: c.SettersSchema,
	}

	// Update the resources with the substitution reference
	inout := &kio.LocalPackageReadWriter{PackagePath: c.ResourcesPath, PackageFileName: c.OpenAPIFileName}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{kio.FilterAll(a)},
		Outputs: []kio.Writer{inout},
	}.Execute()

	if a.Count == 0 {
		fmt.Printf("substitution %s doesn't match any field value in resource configs, "+
			"but creating substitution definition\n", c.Name)
	}
	return err
}

// createMarkersAndRefs takes the input pattern, creates setter/substitution markers
// and corresponding openAPI refs
func (c *SubstitutionCreator) markersAndRefs(substName, pattern string) ([]setters2.Value, error) {
	var values []setters2.Value
	// extract setter name tokens from pattern enclosed in ${}
	re := regexp.MustCompile(`\$\{([^}]*)\}`)
	markers := re.FindAllString(pattern, -1)
	if len(markers) == 0 {
		return nil, errors.Errorf("unable to find setter or substitution names in pattern, " +
			"setter names must be enclosed in ${}")
	}

	for _, marker := range markers {
		name := strings.TrimSuffix(strings.TrimPrefix(marker, "${"), "}")
		if name == substName {
			return nil, fmt.Errorf("setters must have different name than the substitution: %s", name)
		}

		ref, err := spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + name)
		if err != nil {
			return nil, err
		}

		var markerRef string
		subst, _ := openapi.Resolve(&ref, c.SettersSchema)
		// check if the substitution exists with the marker name or fall back to creating setter
		// ref with the name
		if subst != nil {
			markerRef = fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + name
		} else {
			markerRef = fieldmeta.DefinitionsPrefix + fieldmeta.SetterDefinitionPrefix + name
		}

		values = append(
			values,
			setters2.Value{Marker: marker, Ref: markerRef},
		)
	}
	return values, nil
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

	// for each ref in values, check if the setter or substitution already exists, if not create setter
	for _, value := range c.Values {
		// continue if ref is a substitution, as it has already been checked if it exists
		// as part of preRunE
		if strings.Contains(value.Ref, fieldmeta.SubstitutionDefinitionPrefix) {
			fmt.Printf("found a substitution with name %q\n", value.Marker)
			continue
		}
		setterObj, err := y.Pipe(yaml.Lookup(
			// get the setter key from ref. Ex: from #/definitions/io.k8s.cli.setters.image_setter
			// extract io.k8s.cli.setters.image_setter
			"openAPI", "definitions", strings.TrimPrefix(value.Ref, fieldmeta.DefinitionsPrefix)))

		if err != nil {
			return err
		}

		if setterObj == nil {
			name := strings.TrimPrefix(value.Ref, fieldmeta.DefinitionsPrefix+fieldmeta.SetterDefinitionPrefix)
			value := m[value.Marker]
			fmt.Printf("unable to find setter with name %s, creating new setter with value %s\n", name, value)
			sd := setters2.SetterDefinition{
				// get the setter name from ref. Ex: from #/definitions/io.k8s.cli.setters.image_setter
				// extract image_setter
				Name:  name,
				Value: value,
			}
			err := sd.AddToFile(openAPIPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c SubstitutionCreator) checkForCycles(ext *setters2.CliExtension, visited sets.String) error {
	// check if the substitution has already been visited and throw error as cycles
	// are not allowed in nested substitutions
	if visited.Has(ext.Substitution.Name) {
		return errors.Errorf(
			"cyclic substitution detected with name " + ext.Substitution.Name)
	}

	visited.Insert(ext.Substitution.Name)

	// substitute each setter into the pattern to get the new value
	// if substitution references to another substitution, recursively
	// process the nested substitutions to replace the pattern with setter values
	for _, v := range ext.Substitution.Values {
		if v.Ref == "" {
			return errors.Errorf(
				"missing reference on substitution " + ext.Substitution.Name)
		}
		ref, err := spec.NewRef(v.Ref)
		if err != nil {
			return errors.Wrap(err)
		}
		def, err := openapi.Resolve(&ref, c.SettersSchema) // resolve the def to its openAPI def
		if err != nil {
			return errors.Wrap(err)
		}
		defExt, err := setters2.GetExtFromSchema(def) // parse the extension out of the openAPI
		if err != nil {
			return errors.Wrap(err)
		}

		if defExt.Substitution != nil {
			// parse recursively if it reference is substitution
			err := c.checkForCycles(defExt, visited)
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

func (c SubstitutionCreator) validateSubstitutionInfo() error {
	// check if substitution with same name exists and throw error
	ref, err := spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + c.Name)
	if err != nil {
		return err
	}

	subst, _ := openapi.Resolve(&ref, c.SettersSchema)
	// if substitution already exists with the input substitution name, throw error
	if subst != nil {
		return errors.Errorf("substitution with name %q already exists", c.Name)
	}

	// check if setter with same name exists and throw error
	ref, err = spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SetterDefinitionPrefix + c.Name)
	if err != nil {
		return err
	}

	setter, _ := openapi.Resolve(&ref, c.SettersSchema)
	// if setter already exists with input substitution name, throw error
	if setter != nil {
		return errors.Errorf(fmt.Sprintf("setter with name %q already exists, "+
			"substitution and setter can't have same name", c.Name))
	}

	return nil
}
