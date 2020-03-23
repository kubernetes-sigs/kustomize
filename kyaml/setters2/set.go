// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
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
}

// Filter implements Set as a yaml.Filter
func (s *Set) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	return object, accept(s, object)
}

// visitSequence will perform setters for sequences
func (s *Set) visitSequence(object *yaml.RNode, p string, schema *openapi.ResourceSchema) error {
	ext, err := getExtFromComment(schema)
	if err != nil {
		return err
	}
	if ext == nil || ext.Setter == nil || ext.Setter.Name != s.Name ||
		len(ext.Setter.ListValues) == 0 {
		// setter was not invoked for this sequence
		return nil
	}
	s.Count++

	// set the values on the sequences
	var elements []*yaml.Node
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
	if s.set(object, ext, schema.Schema) {
		s.Count++
		return nil
	}

	// perform a substitution of the field if it matches
	sub, err := s.substitute(object, ext, schema.Schema)
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
func (s *Set) substitute(field *yaml.RNode, ext *cliExtension, _ *spec.Schema) (bool, error) {
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

		if val, found := subSetter.Setter.EnumValues[subSetter.Setter.Value]; found {
			// the setter has an enum-map.  we should replace the marker with the
			// enum value looked up from the map rather than the enum key
			p = strings.ReplaceAll(p, v.Marker, val)
		} else {
			// substitute the setters current value into the substitution pattern
			p = strings.ReplaceAll(p, v.Marker, subSetter.Setter.Value)
		}

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

	// substitutions are always strings
	field.YNode().Tag = yaml.StringTag

	return true, nil
}

// set applies the value from ext to field if its name matches s.Name
func (s *Set) set(field *yaml.RNode, ext *cliExtension, sch *spec.Schema) bool {
	// check full setter
	if ext.Setter == nil || ext.Setter.Name != s.Name {
		return false
	}

	// TODO(pwittrock): validate the field value

	if val, found := ext.Setter.EnumValues[ext.Setter.Value]; found {
		// the setter has an enum-map.  we should replace the marker with the
		// enum value looked up from the map rather than the enum key
		field.YNode().Value = val
		return true
	}

	// this has a full setter, set its value
	field.YNode().Value = ext.Setter.Value

	// format the node so it is quoted if it is a string
	yaml.FormatNonStringStyle(field.YNode(), *sch)
	return true
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
	key := SetterDefinitionPrefix + s.Name
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
