// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldmeta

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FieldMeta contains metadata that may be attached to fields as comments
type FieldMeta struct {
	Schema spec.Schema

	Extensions XKustomize
}

type XKustomize struct {
	SetBy               string               `yaml:"setBy,omitempty" json:"setBy,omitempty"`
	PartialFieldSetters []PartialFieldSetter `yaml:"partialSetters,omitempty" json:"partialSetters,omitempty"`
	FieldSetter         *PartialFieldSetter  `yaml:"setter,omitempty" json:"setter,omitempty"`
}

// PartialFieldSetter defines how to set part of a field rather than the full field
// value.  e.g. the tag part of an image field
type PartialFieldSetter struct {
	// Name is the name of this setter.
	Name string `yaml:"name" json:"name"`

	// Value is the current value that has been set.
	Value string `yaml:"value" json:"value"`
}

// Read reads the FieldMeta from a node
func (fm *FieldMeta) Read(n *yaml.RNode) error {
	if n.YNode().LineComment != "" {
		v := strings.TrimLeft(n.YNode().LineComment, "#")
		// if it doesn't Unmarshal that is fine, it means there is no metadata
		// other comments are valid, they just don't parse

		// TODO: consider most sophisticated parsing techniques similar to what is used
		// for go struct tags.
		if err := fm.Schema.UnmarshalJSON([]byte(v)); err != nil {
			// note: don't return an error if the comment isn't a fieldmeta struct
			return nil
		}
		fe := fm.Schema.VendorExtensible.Extensions["x-kustomize"]
		if fe == nil {
			return nil
		}
		b, err := json.Marshal(fe)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, &fm.Extensions)
	}
	return nil
}

// Write writes the FieldMeta to a node
func (fm *FieldMeta) Write(n *yaml.RNode) error {
	fm.Schema.VendorExtensible.AddExtension("x-kustomize", fm.Extensions)
	b, err := json.Marshal(fm.Schema)
	if err != nil {
		return err
	}
	n.YNode().LineComment = string(b)
	return nil
}

// FieldValueType defines the type of input to register
type FieldValueType string

const (
	// String defines a string flag
	String FieldValueType = "string"
	// Bool defines a bool flag
	Bool = "boolean"
	// Int defines an int flag
	Int = "integer"
)

func (it FieldValueType) String() string {
	if it == "" {
		return "string"
	}
	return string(it)
}

func (it FieldValueType) Validate(value string) error {
	switch it {
	case Int:
		if _, err := strconv.Atoi(value); err != nil {
			return errors.WrapPrefixf(err, "value must be an int")
		}
	case Bool:
		if _, err := strconv.ParseBool(value); err != nil {
			return errors.WrapPrefixf(err, "value must be a bool")
		}
	}
	return nil
}

func (it FieldValueType) Tag() string {
	switch it {
	case String:
		return "!!str"
	case Bool:
		return "!!bool"
	case Int:
		return "!!int"
	}
	return ""
}

func (it FieldValueType) TagForValue(value string) string {
	switch it {
	case String:
		return "!!str"
	case Bool:
		if _, err := strconv.ParseBool(string(it)); err != nil {
			return ""
		}
		return "!!bool"
	case Int:
		if _, err := strconv.ParseInt(string(it), 0, 32); err != nil {
			return ""
		}
		return "!!int"
	}
	return ""
}
