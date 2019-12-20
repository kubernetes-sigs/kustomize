// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldmeta

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FieldMeta contains metadata that may be attached to fields as comments
type FieldMeta struct {
	// Substitutions are substitutions that may be performed against this field
	Substitutions []Substitution `yaml:"substitutions,omitempty" json:"substitutions,omitempty"`
	// OwnedBy records the owner of this field
	OwnedBy string `yaml:"setBy,omitempty" json:"setBy,omitempty"`
	// DefaultedBy records that this field was default, but may be changed by other owners
	DefaultedBy string `yaml:"defaultedBy,omitempty" json:"defaultedBy,omitempty"`
	// Description is a description of the current field value, e.g. why it was set
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Type is the type of the field value
	Type FieldValueType `yaml:"type,omitempty" json:"type,omitempty"`
}

// Substitution defines a substitution that may be performed against the field
type Substitution struct {
	// Name is the name of the substitution and read by tools
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Marker is the marker used for replacement
	Marker string `yaml:"marker,omitempty" json:"marker,omitempty"`
	// Value is the current value that has been substituted for the Marker
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

// Read reads the FieldMeta from a node
func (fm *FieldMeta) Read(n *yaml.RNode) error {
	if n.YNode().LineComment != "" {
		v := strings.TrimLeft(n.YNode().LineComment, "#")
		// if it doesn't Unmarshal that is fine, it means there is no metadata
		// other comments are valid, they just don't parse
		d := yaml.NewDecoder(bytes.NewBuffer([]byte(v)))
		d.KnownFields(false)
		_ = d.Decode(fm)
	}
	return nil
}

// Write writes the FieldMeta to a node
func (fm *FieldMeta) Write(n *yaml.RNode) error {
	b, err := json.Marshal(fm)
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
	Bool = "bool"
	// Float defines a float flag
	Float = "float"
	// Int defines an int flag
	Int = "int"
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
	case Float:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return errors.WrapPrefixf(err, "value must be a float")
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
	case Float:
		return "!!float"
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
	case Float:
		if _, err := strconv.ParseFloat(string(it), 64); err != nil {
			return ""
		}
		return "!!float"
	}
	return ""
}
