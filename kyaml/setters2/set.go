// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	goyaml "gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Set sets resource field values from an OpenAPI setter
type Set struct {
	// Name is the name of the setter to set on the object.  i.e. matches the x-k8s-cli.setter.name
	// of the setter that should have its value applied to fields which reference it.
	Name string

	// Count is the number of fields that were updated by calling Filter
	Count int

	// SetAll if set to true will set all setters regardless of name
	SetAll bool

	// Type is the yaml type that will be set on the node. If not provided,
	// the value will be used to infer the type. The provided type must have
	// yaml type format, so one of !!str, !!bool, !!float, !!int and !!null.
	Type string
}

// Filter implements Set as a yaml.Filter
func (s *Set) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	return object, accept(s, object)
}

// isMatch returns true if the setter with name should have the field
// value set
func (s *Set) isMatch(name string) bool {
	return s.SetAll || s.Name == name
}

func (s *Set) visitMapping(object *yaml.RNode, p string, _ *openapi.ResourceSchema) error {
	return nil
}

// visitSequence will perform setters for sequences
func (s *Set) visitSequence(object *yaml.RNode, p string, schema *openapi.ResourceSchema) error {
	ext, err := getExtFromComment(schema)
	if err != nil {
		return err
	}
	if ext == nil || ext.Setter == nil || !s.isMatch(ext.Setter.Name) ||
		len(ext.Setter.ListValues) == 0 {
		// setter was not invoked for this sequence
		return nil
	}
	s.Count++

	// set the values on the sequences
	var elements []*yaml.Node
	if len(ext.Setter.ListValues) > 0 {
		if err := validateAgainstSchema(ext, schema.Schema); err != nil {
			return err
		}
	}
	for i := range ext.Setter.ListValues {
		v := ext.Setter.ListValues[i]
		n := yaml.NewScalarRNode(v).YNode()
		n.Style = yaml.DoubleQuotedStyle
		elements = append(elements, n)
	}
	object.YNode().Content = elements
	object.YNode().Style = yaml.FoldedStyle
	return nil
}

// visitScalar
func (s *Set) visitScalar(object *yaml.RNode, p string, schema *openapi.ResourceSchema) error {
	// get the openAPI for this field describing how to apply the setter
	ext, err := getExtFromComment(schema)
	if err != nil {
		return err
	}
	if ext == nil {
		return nil
	}

	// perform a direct set of the field if it matches
	ok, err := s.set(object, ext, schema.Schema)
	if err != nil {
		return err
	}
	if ok {
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
func (s *Set) substitute(field *yaml.RNode, ext *CliExtension) (bool, error) {
	// check partial setters to see if they contain the setter as part of a
	// substitution
	if ext.Substitution == nil {
		return false, nil
	}

	// track the visited nodes to detect cycles in nested substitutions
	visited := sets.String{}

	// nameMatch indicates if the input substitution depends on the specified setter,
	// the substitution in ext is parsed recursively and if the setter in Set is hit while
	// parsing, it indicates the match
	nameMatch := false

	res, err := s.substituteUtil(ext, visited, &nameMatch)
	if err != nil {
		return false, err
	}

	if !nameMatch {
		// doesn't depend on the setter, don't modify its value
		return false, nil
	}

	field.YNode().Value = res

	// substitutions are always strings
	field.YNode().Tag = yaml.StringTag

	return true, nil
}

// substituteUtil recursively parses nested substitutions in ext and sets the setter value
// returns error if cyclic substitution is detected or any other unexpected errors
func (s *Set) substituteUtil(ext *CliExtension, visited sets.String, nameMatch *bool) (string, error) {
	// check if the substitution has already been visited and throw error as cycles
	// are not allowed in nested substitutions
	if visited.Has(ext.Substitution.Name) {
		return "", errors.Errorf(
			"cyclic substitution detected with name " + ext.Substitution.Name)
	}

	visited.Insert(ext.Substitution.Name)
	pattern := ext.Substitution.Pattern

	// substitute each setter into the pattern to get the new value
	// if substitution references to another substitution, recursively
	// process the nested substitutions to replace the pattern with setter values
	for _, v := range ext.Substitution.Values {
		if v.Ref == "" {
			return "", errors.Errorf(
				"missing reference on substitution " + ext.Substitution.Name)
		}
		ref, err := spec.NewRef(v.Ref)
		if err != nil {
			return "", errors.Wrap(err)
		}
		def, err := openapi.Resolve(&ref) // resolve the def to its openAPI def
		if err != nil {
			return "", errors.Wrap(err)
		}
		defExt, err := GetExtFromSchema(def) // parse the extension out of the openAPI
		if err != nil {
			return "", errors.Wrap(err)
		}

		if defExt.Substitution != nil {
			// parse recursively if it reference is substitution
			substVal, err := s.substituteUtil(defExt, visited, nameMatch)
			if err != nil {
				return "", err
			}
			pattern = strings.ReplaceAll(pattern, v.Marker, substVal)
			continue
		}

		// if code reaches this point, this is a setter, so validate the setter schema
		if err := validateAgainstSchema(defExt, def); err != nil {
			return "", err
		}

		if s.isMatch(defExt.Setter.Name) {
			// the substitution depends on the specified setter
			*nameMatch = true
		}

		if val, found := defExt.Setter.EnumValues[defExt.Setter.Value]; found {
			// the setter has an enum-map.  we should replace the marker with the
			// enum value looked up from the map rather than the enum key
			pattern = strings.ReplaceAll(pattern, v.Marker, val)
		} else {
			pattern = strings.ReplaceAll(pattern, v.Marker, defExt.Setter.Value)
		}
	}

	return pattern, nil
}

// set applies the value from ext to field if its name matches s.Name
func (s *Set) set(field *yaml.RNode, ext *CliExtension, sch *spec.Schema) (bool, error) {
	// check full setter
	if ext.Setter == nil || !s.isMatch(ext.Setter.Name) {
		return false, nil
	}

	if err := validateAgainstSchema(ext, sch); err != nil {
		return false, err
	}

	if val, found := ext.Setter.EnumValues[ext.Setter.Value]; found {
		// the setter has an enum-map.  we should replace the marker with the
		// enum value looked up from the map rather than the enum key
		field.YNode().Value = val
		return true, nil
	}

	var yamlType string
	if s.Type != "" {
		// If a specific type is provided, we just use it instead of trying
		// to infer the type from the value.
		yamlType = s.Type
	} else {
		var err error
		yamlType, _, err = findTypeForField(ext.Setter.Value, field, sch)
		if err != nil {
			return false, err
		}
	}

	// Set the tag of the node to the desired yamlType. The value here
	// might be an empty string, in which case the yaml library will decide
	// on the appropriate value.
	field.YNode().Tag = yamlType

	// this has a full setter, set its value
	field.YNode().Value = ext.Setter.Value

	return true, nil
}

// findTypeForField attempts to infer the appropriate type for the yaml
// field based on the provided value and the schema. If it is able to
// find a type, it will return it as the first return value and the second
// return value will be true. If it is unable to determine a type, it will
// return an empty string and the second return value will be false.
func findTypeForField(newValue string, _ *yaml.RNode, sch *spec.Schema) (string, bool, error) {
	if len(sch.Type) > 0 {
		// If a type is specified in the schema, use it to find the type. We
		// might not be able to find a mapping from the openapi type in the
		// schema to a yaml type, in which case we return an empty string for
		// the yamlType.
		yamlType, found := yaml.YAMLTypeForOpenAPIType(sch.Type[0])
		return yamlType, found, nil
	}

	yamlType, err := findPreferredTypeForValue(newValue)
	if err != nil {
		return "", false, err
	}
	// Only allow the type of a field to be !!null if the schema
	// defines the field as nullable. Otherwise just treat the value
	// as a string.
	if yamlType == yaml.NullTag && !sch.Nullable {
		return yaml.StringTag, true, nil
	}
	return yamlType, true, nil
}

// findPreferredTypeForValue decides the yaml type for a value based on
// two steps:
// First, parse the value as yaml and see which type the underlying library
// decides on. Then, try to look up the corresponding openapi type and
// verify that the provided value is valid for the type in openapi. If it is,
// we return the yaml type from the library. If not, we return a string type.
// This makes sure that values like "yes", which is parsed as a bool value in
// yaml but is not a valid bool value for openapi, is given a string type.
func findPreferredTypeForValue(value string) (string, error) {
	node, err := yaml.Parse(value)
	if err != nil {
		return "", err
	}
	// Get the type assigned to the value by the underlying yaml library.
	yamlType := node.YNode().Tag
	openAPIType, found := yaml.OpenAPITypeForYALMType(yamlType)
	// If we don't have a mapping from the yaml type to an openapi type,
	// we just return the yaml type directly.
	if !found {
		return yamlType, nil
	}

	sc := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{openAPIType},
		},
	}
	var input interface{}
	err = goyaml.Unmarshal([]byte(value), &input)
	if err != nil {
		return "", err
	}
	err = validate.AgainstSchema(&sc, input, strfmt.Default)
	// if the value doesn't validate against an openapi schema with the
	// openapi type, we return the string yaml type.
	if err != nil {
		return yaml.StringTag, nil
	}
	return yamlType, nil
}

// validateAgainstSchema validates the input setter value against user provided
// openAI schema
func validateAgainstSchema(ext *CliExtension, sch *spec.Schema) error {
	sc := spec.Schema{}
	sc.Properties = map[string]spec.Schema{}
	sc.Properties[ext.Setter.Name] = *sch

	var inputYAML string
	if len(ext.Setter.ListValues) > 0 {
		// tmplText contains the template we will use to produce a yaml
		// document that we can use for validation.
		var tmplText string
		if sch.Items != nil && sch.Items.Schema != nil &&
			sch.Items.Schema.Type.Contains("string") {
			// If string is one of the legal types for the value, we
			// output it with quotes in the yaml document to make sure it
			// is later parsed as a string.
			tmplText = `{{.key}}:{{block "list" .values}}{{"\n"}}{{range .}}{{printf "- %q\n" .}}{{end}}{{end}}`
		} else {
			// If string is not specifically set as the type, we just
			// let the yaml unmarshaller detect the correct type. Thus, we
			// do not add quotes around the value.
			tmplText = `{{.key}}:{{block "list" .values}}{{"\n"}}{{range .}}{{println "-" .}}{{end}}{{end}}`
		}

		tmpl, err := template.New("validator").Parse(tmplText)
		if err != nil {
			return err
		}
		var builder strings.Builder
		err = tmpl.Execute(&builder, map[string]interface{}{
			"key":    ext.Setter.Name,
			"values": ext.Setter.ListValues,
		})
		if err != nil {
			return err
		}
		inputYAML = builder.String()
	} else {
		var format string
		// Only add quotes around the value is string is one of the
		// types in the schema.
		if sch.Type.Contains("string") {
			format = "%s: \"%s\""
		} else {
			format = "%s: %s"
		}
		inputYAML = fmt.Sprintf(format, ext.Setter.Name, ext.Setter.Value)
	}

	input := map[string]interface{}{}
	err := goyaml.Unmarshal([]byte(inputYAML), &input)
	if err != nil {
		return err
	}
	err = validate.AgainstSchema(&sc, input, strfmt.Default)
	if err != nil {
		return errors.Errorf("The input value doesn't validate against provided OpenAPI schema: %v\n", err.Error())
	}
	return nil
}

// SetOpenAPI updates a setter value
type SetOpenAPI struct {
	// Name is the name of the setter to add
	Name string `yaml:"name"`
	// Value is the current value of the setter
	Value string `yaml:"value"`

	// ListValue is the current value for a list of items
	ListValues []string `yaml:"listValue"`

	Description string `yaml:"description"`

	SetBy string `yaml:"setBy"`
}

// UpdateFile updates the OpenAPI definitions in a file with the given setter value.
func (s SetOpenAPI) UpdateFile(path string) error {
	return yaml.UpdateFile(s, path)
}

func (s SetOpenAPI) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	key := fieldmeta.SetterDefinitionPrefix + s.Name
	oa, err := object.Pipe(yaml.Lookup("openAPI", "definitions", key))
	if err != nil {
		return nil, err
	}
	if oa == nil {
		return nil, errors.Errorf("no setter %s found", s.Name)
	}
	def, err := oa.Pipe(yaml.Lookup("x-k8s-cli", "setter"))
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, errors.Errorf("no setter %s found", s.Name)
	}

	// record the OpenAPI type for the setter
	var t string
	if n := oa.Field("type"); n != nil {
		t = n.Value.YNode().Value
	}

	// if the setter contains an enumValues map, then ensure the set value appears
	// as a key in the map
	if values, err := def.Pipe(
		yaml.Lookup("enumValues")); err != nil {
		// error looking up the enumValues
		return nil, err
	} else if values != nil {
		// contains enumValues map -- validate the set value against the map entries

		// get the enumValues keys
		fields, err := values.Fields()
		if err != nil {
			return nil, err
		}

		// search for the user provided value in the set of allowed values
		var match bool
		for i := range fields {
			if fields[i] == s.Value {
				// found a match, we are good
				match = true
				break
			}
		}
		if !match {
			// no match found -- provide an informative error to the user
			return nil, errors.Errorf("%s does not match the possible values for %s: [%s]",
				s.Value, s.Name, strings.Join(fields, ","))
		}
	}

	v := yaml.NewScalarRNode(s.Value)
	// values are always represented as strings the OpenAPI
	// since the are unmarshalled into strings.  Use double quote style to
	// ensure this consistently.
	v.YNode().Tag = yaml.StringTag
	v.YNode().Style = yaml.DoubleQuotedStyle

	if t != "array" {
		// set a scalar value
		if err := def.PipeE(&yaml.FieldSetter{Name: "value", Value: v}); err != nil {
			return nil, err
		}
	} else {
		// set a list value
		if err := def.PipeE(&yaml.FieldClearer{Name: "value"}); err != nil {
			return nil, err
		}
		// create the list values
		var elements []*yaml.Node
		n := yaml.NewScalarRNode(s.Value).YNode()
		n.Tag = yaml.StringTag
		n.Style = yaml.DoubleQuotedStyle
		elements = append(elements, n)
		for i := range s.ListValues {
			v := s.ListValues[i]
			n := yaml.NewScalarRNode(v).YNode()
			n.Style = yaml.DoubleQuotedStyle
			elements = append(elements, n)
		}
		l := yaml.NewRNode(&yaml.Node{
			Kind:    yaml.SequenceNode,
			Content: elements,
		})

		def.YNode().Style = yaml.FoldedStyle
		if err := def.PipeE(&yaml.FieldSetter{Name: "listValues", Value: l}); err != nil {
			return nil, err
		}
	}

	if err := def.PipeE(&yaml.FieldSetter{Name: "setBy", StringValue: s.SetBy}); err != nil {
		return nil, err
	}

	if err := def.PipeE(&yaml.FieldSetter{Name: "isSet", StringValue: "true"}); err != nil {
		return nil, err
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

// SetAll applies the set filter for all yaml nodes and only returns the nodes whose
// corresponding file has at least one node with input setter
func SetAll(s *Set) kio.Filter {
	return kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		filesToUpdate := sets.String{}
		// for each node record the set fields count before and after filter is applied and
		// store the corresponding file paths if there is an increment in setters count
		for i := range nodes {
			preCount := s.Count
			_, err := s.Filter(nodes[i])
			if err != nil {
				return nil, errors.Wrap(err)
			}
			if s.Count > preCount {
				path, _, err := kioutil.GetFileAnnotations(nodes[i])
				if err != nil {
					return nil, errors.Wrap(err)
				}
				filesToUpdate.Insert(path)
			}
		}
		var nodesInUpdatedFiles []*yaml.RNode
		// return only the nodes whose corresponding file has at least one node with input setter
		for i := range nodes {
			path, _, err := kioutil.GetFileAnnotations(nodes[i])
			if err != nil {
				return nil, errors.Wrap(err)
			}
			if filesToUpdate.Has(path) {
				nodesInUpdatedFiles = append(nodesInUpdatedFiles, nodes[i])
			}
		}
		return nodesInUpdatedFiles, nil
	})
}
