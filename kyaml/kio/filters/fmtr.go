// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package yamlfmt contains libraries for formatting yaml files containing
// Kubernetes Resource configuration.
//
// Yaml files are formatted by:
// - Sorting fields and map values
// - Sorting unordered lists for whitelisted types
// - Applying a canonical yaml Style
//
// Fields are ordered using a relative ordering applied to commonly
// encountered Resource fields.  All Resources,  including non-builtin
// Resources such as CRDs, share the same field precedence.
//
// Fields that do not appear in the explicit ordering are ordered
// lexicographically.
//
// A subset of well known known unordered lists are sorted by element field
// values.
package filters

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FormatInput returns the formatted input.
func FormatInput(input io.Reader) (*bytes.Buffer, error) {
	buff := &bytes.Buffer{}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: input}},
		Filters: []kio.Filter{FormatFilter{}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: buff}},
	}.Execute()

	return buff, err
}

// FormatFileOrDirectory reads the file or directory and formats each file's
// contents by writing it back to the file.
func FormatFileOrDirectory(path string) error {
	return kio.Pipeline{
		Inputs: []kio.Reader{kio.LocalPackageReader{
			PackagePath: path,
		}},
		Filters: []kio.Filter{FormatFilter{}},
		Outputs: []kio.Writer{kio.LocalPackageWriter{PackagePath: path}},
	}.Execute()
}

type FormatFilter struct{}

var _ kio.Filter = FormatFilter{}

func (f FormatFilter) Filter(slice []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range slice {
		kindNode, err := slice[i].Pipe(yaml.Get("kind"))
		if err != nil {
			return nil, err
		}
		if kindNode == nil {
			continue
		}
		apiVersionNode, err := slice[i].Pipe(yaml.Get("apiVersion"))
		if err != nil {
			return nil, err
		}
		if apiVersionNode == nil {
			continue
		}
		kind, apiVersion := kindNode.YNode().Value, apiVersionNode.YNode().Value
		err = (&formatter{apiVersion: apiVersion, kind: kind}).fmtNode(slice[i].YNode(), "")
		if err != nil {
			return nil, err
		}
	}
	return slice, nil
}

type formatter struct {
	apiVersion string
	kind       string
}

// fmtNode recursively formats the Document Contents.
func (f *formatter) fmtNode(n *yaml.Node, path string) error {
	// sort the order of mapping fields
	if n.Kind == yaml.MappingNode {
		sort.Sort(sortedMapContents(*n))
	}

	// sort the order of sequence elements if it is whitelisted
	if n.Kind == yaml.SequenceNode {
		if yaml.WhitelistedListSortKinds.Has(f.kind) &&
			yaml.WhitelistedListSortApis.Has(f.apiVersion) {
			if sortField, found := yaml.WhitelistedListSortFields[path]; found {
				sort.Sort(sortedSeqContents{Node: *n, sortField: sortField})
			}
		}
	}
	for i := range n.Content {
		p := path
		if n.Kind == yaml.MappingNode && i%2 == 1 {
			p = fmt.Sprintf("%s.%s", path, n.Content[i-1].Value)
		}
		err := f.fmtNode(n.Content[i], p)
		if err != nil {
			return err
		}
	}
	return nil
}

// sortedMapContents sorts the Contents field of a MappingNode by the field names using a statically
// defined field precedence, and falling back on lexicographical sorting
type sortedMapContents yaml.Node

func (s sortedMapContents) Len() int {
	return len(s.Content) / 2
}
func (s sortedMapContents) Swap(i, j int) {
	// yaml MappingNode Contents are a list of field names followed by
	// field values, rather than a list of field <name, value> pairs.
	// increment.
	//
	// e.g. ["field1Name", "field1Value", "field2Name", "field2Value"]
	iFieldNameIndex := i * 2
	jFieldNameIndex := j * 2
	iFieldValueIndex := iFieldNameIndex + 1
	jFieldValueIndex := jFieldNameIndex + 1

	// swap field names
	s.Content[iFieldNameIndex], s.Content[jFieldNameIndex] =
		s.Content[jFieldNameIndex], s.Content[iFieldNameIndex]

	// swap field values
	s.Content[iFieldValueIndex], s.Content[jFieldValueIndex] = s.
		Content[jFieldValueIndex], s.Content[iFieldValueIndex]
}
func (s sortedMapContents) Less(i, j int) bool {
	iFieldNameIndex := i * 2
	jFieldNameIndex := j * 2
	iFieldName := s.Content[iFieldNameIndex].Value
	jFieldName := s.Content[jFieldNameIndex].Value

	// order by their precedence values looked up from the index
	iOrder, foundI := yaml.FieldOrder[iFieldName]
	jOrder, foundJ := yaml.FieldOrder[jFieldName]
	if foundI && foundJ {
		return iOrder < jOrder
	}

	// known fields come before unknown fields
	if foundI {
		return true
	}
	if foundJ {
		return false
	}

	// neither field is known, sort them lexicographically
	return iFieldName < jFieldName
}

// sortedSeqContents sorts the Contents field of a SequenceNode by the value of
// the elements sortField.
// e.g. it will sort spec.template.spec.containers by the value of the container `name` field
type sortedSeqContents struct {
	yaml.Node
	sortField string
}

func (s sortedSeqContents) Len() int {
	return len(s.Content)
}
func (s sortedSeqContents) Swap(i, j int) {
	s.Content[i], s.Content[j] = s.Content[j], s.Content[i]
}
func (s sortedSeqContents) Less(i, j int) bool {
	// primitive lists -- sort by the element's primitive values
	if s.sortField == "" {
		iValue := s.Content[i].Value
		jValue := s.Content[j].Value
		return iValue < jValue
	}

	// map lists -- sort by the element's sortField values
	var iValue, jValue string
	for a := range s.Content[i].Content {
		if a%2 != 0 {
			continue // not a fieldNameIndex
		}
		// locate the index of the sortField field
		if s.Content[i].Content[a].Value == s.sortField {
			// a is the yaml node for the field key, a+1 is the node for the field value
			iValue = s.Content[i].Content[a+1].Value
		}
	}
	for a := range s.Content[j].Content {
		if a%2 != 0 {
			continue // not a fieldNameIndex
		}

		// locate the index of the sortField field
		if s.Content[j].Content[a].Value == s.sortField {
			// a is the yaml node for the field key, a+1 is the node for the field value
			jValue = s.Content[j].Content[a+1].Value
		}
	}

	// compare the field values
	return iValue < jValue
}
