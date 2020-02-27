// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"encoding/json"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type cliExtension struct {
	Setter       *setter       `yaml:"setter,omitempty" json:"setter,omitempty"`
	Substitution *substitution `yaml:"substitution,omitempty" json:"substitution,omitempty"`
}

type setter struct {
	Name       string            `yaml:"name,omitempty" json:"name,omitempty"`
	Value      string            `yaml:"value,omitempty" json:"value,omitempty"`
	EnumValues map[string]string `yaml:"enumValues,omitempty" json:"enumValues,omitempty"`
}

type substitution struct {
	Name    string                        `yaml:"name,omitempty" json:"name,omitempty"`
	Pattern string                        `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Values  []substitutionSetterReference `yaml:"values,omitempty" json:"values,omitempty"`
}

type substitutionSetterReference struct {
	Ref    string `yaml:"ref,omitempty" json:"ref,omitempty"`
	Marker string `yaml:"marker,omitempty" json:"marker,omitempty"`
}

//K8sCliExtensionKey is the name of the OpenAPI field containing the setter extensions
const K8sCliExtensionKey = "x-k8s-cli"

// getExtFromSchema returns the cliExtension openAPI extension if it is present in schema
func getExtFromSchema(schema *spec.Schema) (*cliExtension, error) {
	cep := schema.VendorExtensible.Extensions[K8sCliExtensionKey]
	if cep == nil {
		return nil, nil
	}
	b, err := json.Marshal(cep)
	if err != nil {
		return nil, err
	}
	val := &cliExtension{}
	if err := json.Unmarshal(b, val); err != nil {
		return nil, err
	}
	return val, nil
}

// getExtFromComment returns the cliExtension openAPI extension if it is present as
// a comment on the field.
func getExtFromComment(object *yaml.RNode) (*cliExtension, *spec.Schema, error) {
	// TODO(pwittrock): also use path to the field to get openapi, not just comments
	// parse comment containing the extended openapi for this field
	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(object); err != nil {
		return nil, nil, errors.Wrap(err)
	}
	if fm.Schema.Ref.String() == "" {
		return nil, nil, nil
	}

	// resolve the comment reference to the extended openapi definitions
	r, err := openapi.Resolve(&fm.Schema.Ref)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	if r == nil {
		// no schema found
		// TODO(pwittrock): should this be an error if it doesn't resolve?
		return nil, nil, nil
	}

	// get the cli extension from the openapi (contains setter information)
	ext, err := getExtFromSchema(r)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	return ext, r, nil
}
