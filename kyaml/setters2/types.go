// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"encoding/json"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

type CliExtension struct {
	Setter       *setter       `yaml:"setter,omitempty" json:"setter,omitempty"`
	Substitution *substitution `yaml:"substitution,omitempty" json:"substitution,omitempty"`
}

type setter struct {
	Name       string            `yaml:"name,omitempty" json:"name,omitempty"`
	Value      string            `yaml:"value,omitempty" json:"value,omitempty"`
	ListValues []string          `yaml:"listValues,omitempty" json:"listValues,omitempty"`
	EnumValues map[string]string `yaml:"enumValues,omitempty" json:"enumValues,omitempty"`
	Required   bool              `yaml:"required,omitempty" json:"required,omitempty"`
	IsSet      bool              `yaml:"isSet,omitempty" json:"isSet,omitempty"`
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

// K8sCliExtensionKey is the name of the OpenAPI field containing the setter extensions
const K8sCliExtensionKey = "x-k8s-cli"

// GetExtFromSchema returns the cliExtension openAPI extension if it is present in schema
func GetExtFromSchema(schema *spec.Schema) (*CliExtension, error) {
	cep := schema.VendorExtensible.Extensions[K8sCliExtensionKey]
	if cep == nil {
		return nil, nil
	}
	b, err := json.Marshal(cep)
	if err != nil {
		return nil, err
	}
	val := &CliExtension{}
	if err := json.Unmarshal(b, val); err != nil {
		return nil, err
	}
	return val, nil
}

// getExtFromComment returns the cliExtension openAPI extension if it is present as
// a comment on the field.
func getExtFromComment(schema *openapi.ResourceSchema) (*CliExtension, error) {
	if schema == nil {
		// no schema found
		// TODO(pwittrock): should this be an error if it doesn't resolve?
		return nil, nil
	}

	// get the cli extension from the openapi (contains setter information)
	ext, err := GetExtFromSchema(schema.Schema)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return ext, nil
}
