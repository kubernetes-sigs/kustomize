// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Severity indicates the severity of the Result
type Severity string

const (
	// Error indicates the result is an error.  Will cause the function to exit non-0.
	Error Severity = "error"
	// Warning indicates the result is a warning
	Warning Severity = "warning"
	// Info indicates the result is an informative message
	Info Severity = "info"
)

// ResultItem defines a validation result
type Result struct {
	// Message is a human readable message. This field is required.
	Message string `yaml:"message,omitempty" json:"message,omitempty"`

	// Severity is the severity of this result
	Severity Severity `yaml:"severity,omitempty" json:"severity,omitempty"`

	// ResourceRef is a reference to a resource.
	// Required fields: apiVersion, kind, name.
	ResourceRef *yaml.ResourceIdentifier `yaml:"resourceRef,omitempty" json:"resourceRef,omitempty"`

	// Field is a reference to the field in a resource this result refers to
	Field *Field `yaml:"field,omitempty" json:"field,omitempty"`

	// File references a file containing the resource this result refers to
	File *File `yaml:"file,omitempty" json:"file,omitempty"`

	// Tags is an unstructured key value map stored with a result that may be set
	// by external tools to store and retrieve arbitrary metadata
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// String provides a human-readable message for the result item
func (i Result) String() string {
	identifier := i.ResourceRef
	var idStringList []string
	if identifier != nil {
		if identifier.APIVersion != "" {
			idStringList = append(idStringList, identifier.APIVersion)
		}
		if identifier.Kind != "" {
			idStringList = append(idStringList, identifier.Kind)
		}
		if identifier.Namespace != "" {
			idStringList = append(idStringList, identifier.Namespace)
		}
		if identifier.Name != "" {
			idStringList = append(idStringList, identifier.Name)
		}
	}
	formatString := "[%s]"
	severity := i.Severity
	// We default Severity to Info when converting a result to a message.
	if i.Severity == "" {
		severity = Info
	}
	list := []interface{}{severity}
	if len(idStringList) > 0 {
		formatString += " %s"
		list = append(list, strings.Join(idStringList, "/"))
	}
	if i.Field != nil {
		formatString += " %s"
		list = append(list, i.Field.Path)
	}
	formatString += ": %s"
	list = append(list, i.Message)
	return fmt.Sprintf(formatString, list...)
}

// File references a file containing a resource
type File struct {
	// Path is relative path to the file containing the resource.
	// This field is required.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Index is the index into the file containing the resource
	// (i.e. if there are multiple resources in a single file)
	Index int `yaml:"index,omitempty" json:"index,omitempty"`
}

// Field references a field in a resource
type Field struct {
	// Path is the field path. This field is required.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// CurrentValue is the current field value
	CurrentValue interface{} `yaml:"currentValue,omitempty" json:"currentValue,omitempty"`

	// ProposedValue is the proposed value of the field to fix an issue.
	ProposedValue interface{} `yaml:"proposedValue,omitempty" json:"proposedValue,omitempty"`
}

type Results []*Result

// Error enables Results to be returned as an error
func (e Results) Error() string {
	var msgs []string
	for _, i := range e {
		msgs = append(msgs, i.String())
	}
	return strings.Join(msgs, "\n\n")
}

// ExitCode provides the exit code based on the result's severity
func (e Results) ExitCode() int {
	for _, i := range e {
		if i.Severity == Error {
			return 1
		}
	}
	return 0
}
