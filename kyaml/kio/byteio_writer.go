// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio

import (
	"encoding/json"
	"io"
	"path/filepath"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ByteWriter writes ResourceNodes to bytes.
type ByteWriter struct {
	// Writer is where ResourceNodes are encoded.
	Writer io.Writer

	// KeepReaderAnnotations if set will keep the Reader specific annotations when writing
	// the Resources, otherwise they will be cleared.
	KeepReaderAnnotations bool

	// ClearAnnotations is a list of annotations to clear when writing the Resources.
	ClearAnnotations []string

	// Style is a style that is set on the Resource Node Document.
	Style yaml.Style

	// FunctionConfig is the function config for an ResourceList.  If non-nil
	// wrap the results in an ResourceList.
	FunctionConfig *yaml.RNode

	Results *yaml.RNode

	// WrappingKind if set will cause ByteWriter to wrap the Resources in
	// an 'items' field in this kind.  e.g. if WrappingKind is 'List',
	// ByteWriter will wrap the Resources in a List .items field.
	WrappingKind string

	// WrappingAPIVersion is the apiVersion for WrappingKind
	WrappingAPIVersion string

	// Sort if set, will cause ByteWriter to sort the the nodes before writing them.
	Sort bool
}

var _ Writer = ByteWriter{}

func (w ByteWriter) Write(nodes []*yaml.RNode) error {
	yaml.DoSerializationHacksOnNodes(nodes)
	if w.Sort {
		if err := kioutil.SortNodes(nodes); err != nil {
			return errors.Wrap(err)
		}
	}

	// Check if the output is a single JSON file so we can force JSON encoding in that case.
	// YAML flow style encoding may not be compatible because of unquoted strings and newlines
	// introduced by the YAML marshaller in long string values. These newlines are insignificant
	// when interpreted as YAML but invalid when interpreted as JSON.
	jsonEncodeSingleNode := false
	if w.WrappingKind == "" && len(nodes) == 1 {
		if path, _, _ := kioutil.GetFileAnnotations(nodes[0]); path != "" {
			filename := filepath.Base(path)
			for _, glob := range JSONMatch {
				if match, _ := filepath.Match(glob, filename); match {
					jsonEncodeSingleNode = true
					break
				}
			}
		}
	}

	for i := range nodes {
		// clean resources by removing annotations set by the Reader
		if !w.KeepReaderAnnotations {
			_, err := nodes[i].Pipe(yaml.ClearAnnotation(kioutil.IndexAnnotation))
			if err != nil {
				return errors.Wrap(err)
			}
		}
		for _, a := range w.ClearAnnotations {
			_, err := nodes[i].Pipe(yaml.ClearAnnotation(a))
			if err != nil {
				return errors.Wrap(err)
			}
		}

		if err := yaml.ClearEmptyAnnotations(nodes[i]); err != nil {
			return err
		}

		if w.Style != 0 {
			nodes[i].YNode().Style = w.Style
		}
	}

	if jsonEncodeSingleNode {
		encoder := json.NewEncoder(w.Writer)
		encoder.SetIndent("", "  ")
		return errors.Wrap(encoder.Encode(nodes[0]))
	}

	encoder := yaml.NewEncoder(w.Writer)
	defer encoder.Close()
	// don't wrap the elements
	if w.WrappingKind == "" {
		for i := range nodes {
			if err := encoder.Encode(nodes[i].Document()); err != nil {
				return err
			}
		}
		return nil
	}
	// wrap the elements in a list
	items := &yaml.Node{Kind: yaml.SequenceNode}
	list := &yaml.Node{
		Kind:  yaml.MappingNode,
		Style: w.Style,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "apiVersion"},
			{Kind: yaml.ScalarNode, Value: w.WrappingAPIVersion},
			{Kind: yaml.ScalarNode, Value: "kind"},
			{Kind: yaml.ScalarNode, Value: w.WrappingKind},
			{Kind: yaml.ScalarNode, Value: "items"}, items,
		}}
	if w.FunctionConfig != nil {
		list.Content = append(list.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "functionConfig"},
			w.FunctionConfig.YNode())
	}
	if w.Results != nil {
		list.Content = append(list.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "results"},
			w.Results.YNode())
	}
	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{list}}
	for i := range nodes {
		items.Content = append(items.Content, nodes[i].YNode())
	}
	err := encoder.Encode(doc)
	yaml.UndoSerializationHacksOnNodes(nodes)
	return err
}
