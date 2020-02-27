// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Add creates or updates setter or substitution references from resource fields.
// Requires that at least one of FieldValue and FieldName have been set.
type Add struct {
	// FieldValue if set will add the OpenAPI reference to fields if they have this value.
	// Optional.  If unspecified match all field values.
	FieldValue string

	// FieldName if set will add the OpenAPI reference to fields with this name or path
	// FieldName may be the full name of the field, full path to the field, or the path suffix.
	// e.g. all of the following would match spec.template.spec.containers.image --
	// [image, containers.image, spec.containers.image, template.spec.containers.image,
	//  spec.template.spec.containers.image]
	// Optional.  If unspecified match all field names.
	FieldName string

	// Ref is the OpenAPI reference to set on the matching fields as a comment.
	Ref string
}

// Filter implements yaml.Filter
func (a *Add) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	if a.FieldName == "" && a.FieldValue == "" {
		return nil, errors.Errorf("must specify either fieldName or fieldValue")
	}
	if a.Ref == "" {
		return nil, errors.Errorf("must specify ref")
	}
	return object, accept(a, object)
}

// visitScalar implements visitor
// visitScalar will set the field metadata on each scalar field whose name + value match
func (a *Add) visitScalar(object *yaml.RNode, p string) error {
	// check if the field matches
	if a.FieldName != "" && !strings.HasSuffix(p, a.FieldName) {
		return nil
	}
	if a.FieldValue != "" && a.FieldValue != object.YNode().Value {
		return nil
	}

	// read the field metadata
	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(object); err != nil {
		return err
	}

	// create the ref on the field metadata
	r, err := spec.NewRef(a.Ref)
	if err != nil {
		return err
	}
	fm.Schema.Ref = r

	// write the field metadata
	if err := fm.Write(object); err != nil {
		return err
	}
	return nil
}

const (
	// CLIDefinitionsPrefix is the prefix for cli definition keys.
	CLIDefinitionsPrefix = "io.k8s.cli."

	// SetterDefinitionPrefix is the prefix for setter definition keys.
	SetterDefinitionPrefix = CLIDefinitionsPrefix + "setters."

	// SubstitutionDefinitionPrefix is the prefix for substitution definition keys.
	SubstitutionDefinitionPrefix = CLIDefinitionsPrefix + "substitutions."

	// DefinitionsPrefix is the prefix used to reference definitions in the OpenAPI
	DefinitionsPrefix = "#/definitions/"
)

// SetterDefinition may be used to update a files OpenAPI definitions with a new setter.
type SetterDefinition struct {
	// Name is the name of the setter to create or update.
	Name string `yaml:"name"`

	// Value is the value of the setter.
	Value string `yaml:"value"`

	// SetBy is the person or role that last set the value.
	SetBy string `yaml:"setBy,omitempty"`

	// Description is a description of the value.
	Description string `yaml:"description,omitempty"`

	// Count is the number of fields set by this setter.
	Count int `yaml:"count,omitempty"`

	// Type is the type of the setter value.
	Type string `yaml:"type,omitempty"`

	// EnumValues is a map of possible setter values to actual field values.
	// If EnumValues is specified, then the value set the by user 1) MUST
	// be present in the enumValues map as a key, and 2) the map entry value
	// MUST be used as the value to set in the configuration (rather than the key)
	// Example -- may be used for t-shirt sizing values by allowing cpu to be
	// set to small, medium or large, and then mapping these values to cpu values -- 0.5, 2, 8
	EnumValues map[string]string `yaml:"enumValues,omitempty"`
}

func (sd SetterDefinition) AddToFile(path string) error {
	return yaml.UpdateFile(sd, path)
}

func (sd SetterDefinition) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	key := SetterDefinitionPrefix + sd.Name

	def, err := object.Pipe(yaml.LookupCreate(
		yaml.MappingNode, openapi.SupplementaryOpenAPIFieldName, "definitions", key))
	if err != nil {
		return nil, err
	}
	if sd.Description != "" {
		err = def.PipeE(yaml.FieldSetter{Name: "description", StringValue: sd.Description})
		if err != nil {
			return nil, err
		}
		// don't write the description to the extension
		sd.Description = ""
	}

	if sd.Type != "" {
		err = def.PipeE(yaml.FieldSetter{Name: "type", StringValue: sd.Type})
		if err != nil {
			return nil, err
		}
		// don't write the type to the extension
		sd.Type = ""
	}

	ext, err := def.Pipe(yaml.LookupCreate(yaml.MappingNode, K8sCliExtensionKey))
	if err != nil {
		return nil, err
	}

	b, err := yaml.Marshal(sd)
	if err != nil {
		return nil, err
	}
	y, err := yaml.Parse(string(b))
	if err != nil {
		return nil, err
	}

	if err := ext.PipeE(yaml.SetField("setter", y)); err != nil {
		return nil, err
	}

	return object, nil
}

// SetterDefinition may be used to update a files OpenAPI definitions with a new substitution.
type SubstitutionDefinition struct {
	// Name is the name of the substitution to create or update
	Name string `yaml:"name"`

	// Pattern is the substitution pattern into which setter values are substituted
	Pattern string `yaml:"pattern"`

	// Values are setters which are substituted into pattern to produce a field value
	Values []Value `yaml:"values"`
}

type Value struct {
	// Marker is the string marker in pattern that is replace by the referenced setter.
	Marker string `yaml:"marker"`

	// Ref is a reference to a setter to pull the replacement value from.
	Ref string `yaml:"ref"`
}

func (sd SubstitutionDefinition) AddToFile(path string) error {
	return yaml.UpdateFile(sd, path)
}

func (sd SubstitutionDefinition) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	// create the substitution extension value by marshalling the SubstitutionDefinition itself
	b, err := yaml.Marshal(sd)
	if err != nil {
		return nil, err
	}
	sub, err := yaml.Parse(string(b))
	if err != nil {
		return nil, err
	}

	// lookup or create the definition for the substitution
	defKey := SubstitutionDefinitionPrefix + sd.Name
	def, err := object.Pipe(yaml.LookupCreate(
		yaml.MappingNode, openapi.SupplementaryOpenAPIFieldName, "definitions", defKey, "x-k8s-cli"))
	if err != nil {
		return nil, err
	}

	// set the substitution on the definition
	if err := def.PipeE(yaml.SetField("substitution", sub)); err != nil {
		return nil, err
	}

	return object, nil
}
