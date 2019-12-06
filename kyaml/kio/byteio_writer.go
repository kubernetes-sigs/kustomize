// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio

import (
	"io"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Writer writes ResourceNodes to bytes.
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
	if w.Sort {
		if err := kioutil.SortNodes(nodes); err != nil {
			return errors.Wrap(err)
		}
	}

	encoder := yaml.NewEncoder(w.Writer)
	defer encoder.Close()
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

		// TODO(pwittrock): factor this into a a common module for pruning empty values
		_, err := nodes[i].Pipe(yaml.Lookup("metadata"), yaml.FieldClearer{
			Name: "annotations", IfEmpty: true})
		if err != nil {
			return errors.Wrap(err)
		}
		_, err = nodes[i].Pipe(yaml.FieldClearer{Name: "metadata", IfEmpty: true})
		if err != nil {
			return errors.Wrap(err)
		}

		if w.Style != 0 {
			nodes[i].YNode().Style = w.Style
		}
	}

	// don't wrap the elements
	if w.WrappingKind == "" {
		for i := range nodes {
			err := encoder.Encode(nodes[i].Document())
			if err != nil {
				return errors.Wrap(err)
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
	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Style:   yaml.FoldedStyle,
		Content: []*yaml.Node{list}}
	for i := range nodes {
		items.Content = append(items.Content, nodes[i].YNode())
	}
	return errors.Wrap(encoder.Encode(doc))
}
