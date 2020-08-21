// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"regexp"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SearchReplace struct holds the input parameters and results for
// Search and Replace operations on resource configs
type SearchReplace struct {
	// Value is the value of the field to be matched
	Value string

	// ValueRegex is the regex of the field to be matched
	ValueRegex string

	// Path is the path of the field to be matched
	Path string

	// ReplaceLiteral is the value with which the matched field is replaced
	ReplaceLiteral string

	// Count is the number of matches
	Count int

	// Match is the list of matched resource nodes
	Match []*yaml.RNode
}

// Perform performs the search and replace operation on each node in the package path
func (sr *SearchReplace) Perform(resourcesPath string) error {
	inout := &kio.LocalPackageReadWriter{PackagePath: resourcesPath, NoDeleteFiles: true}
	return kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{kio.FilterAll(sr)},
		Outputs: []kio.Writer{inout},
	}.Execute()
}

// Filter parses input node and performs search and replace operation on the node
func (sr *SearchReplace) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	preCount := sr.Count
	err := accept(sr, object)
	// consider the node is matched if the matched field count increases
	if sr.Count > preCount {
		sr.Match = append(sr.Match, object)
	}
	return object, err
}

// visitMapping parses mapping node
func (sr *SearchReplace) visitMapping(object *yaml.RNode, p string, _ *openapi.ResourceSchema) error {
	return nil
}

// visitSequence parses sequence node
func (sr *SearchReplace) visitSequence(object *yaml.RNode, p string, _ *openapi.ResourceSchema) error {
	return nil
}

// visitScalar parses scalar node
func (sr *SearchReplace) visitScalar(object *yaml.RNode, p string, _, _ *openapi.ResourceSchema) error {
	regexMatch, err := sr.regexMatch(object.Document().Value)
	if err != nil {
		return err
	}

	if object.Document().Value == sr.Value || regexMatch {
		sr.Count++
		if sr.ReplaceLiteral != "" {
			object.Document().Value = sr.ReplaceLiteral
		}
	}
	return nil
}

// regexMatch checks if ValueRegex in SearchReplace struct matches with the input
// value, returns error if any
func (sr *SearchReplace) regexMatch(value string) (bool, error) {
	if sr.ValueRegex == "" {
		return false, nil
	}
	re, err := regexp.Compile(sr.ValueRegex)
	if err != nil {
		return false, errors.Wrap(err)
	}
	return re.Match([]byte(value)), nil
}
