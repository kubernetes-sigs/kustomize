// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

// IsYNodeTaggedNull returns true if the node is explicitly tagged Null.
func IsYNodeTaggedNull(n *yaml.Node) bool {
	return n != nil && n.Tag == NodeTagNull
}

// IsYNodeEmptyMap is true if the Node is a non-nil empty map.
func IsYNodeEmptyMap(n *yaml.Node) bool {
	return n != nil && n.Kind == yaml.MappingNode && len(n.Content) == 0
}

// IsYNodeEmptyMap is true if the Node is a non-nil empty sequence.
func IsYNodeEmptySeq(n *yaml.Node) bool {
	return n != nil && n.Kind == yaml.SequenceNode && len(n.Content) == 0
}

// IsYNodeEmptyDoc is true if the node is a Document with no content.
// E.g.: "---\n---"
func IsYNodeEmptyDoc(n *yaml.Node) bool {
	return n.Kind == yaml.DocumentNode && n.Content[0].Tag == NodeTagNull
}

func IsYNodeString(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == NodeTagString
}

// Parser parses values into configuration.
type Parser struct {
	Kind  string `yaml:"kind,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (p Parser) Filter(_ *RNode) (*RNode, error) {
	d := yaml.NewDecoder(bytes.NewBuffer([]byte(p.Value)))
	o := &RNode{value: &yaml.Node{}}
	return o, d.Decode(o.value)
}

// Parse parses a yaml string into an *RNode
func Parse(value string) (*RNode, error) {
	return Parser{Value: value}.Filter(nil)
}

// ReadFile parses a single Resource from a yaml file
func ReadFile(path string) (*RNode, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(b))
}

// WriteFile writes a single Resource to a yaml file
func WriteFile(node *RNode, path string) error {
	out, err := node.String()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(out), 0600)
}

// UpdateFile reads the file at path, applies the filter to it, and write the result back.
// path must contain a exactly 1 resource (YAML).
func UpdateFile(filter Filter, path string) error {
	// Read the yaml
	y, err := ReadFile(path)
	if err != nil {
		return err
	}

	// Update the yaml
	if err := y.PipeE(filter); err != nil {
		return err
	}

	// Write the yaml
	return WriteFile(y, path)
}

// TODO(pwittrock): test this
func GetStyle(styles ...string) Style {
	var style Style
	for _, s := range styles {
		switch s {
		case "TaggedStyle":
			style |= TaggedStyle
		case "DoubleQuotedStyle":
			style |= DoubleQuotedStyle
		case "SingleQuotedStyle":
			style |= SingleQuotedStyle
		case "LiteralStyle":
			style |= LiteralStyle
		case "FoldedStyle":
			style |= FoldedStyle
		case "FlowStyle":
			style |= FlowStyle
		}
	}
	return style
}

// Filter defines a function to manipulate an individual RNode such as by changing
// its values, or returning a field.
//
// When possible, Filters should be serializable to yaml so that they can be described
// declaratively as data.
//
// Analogous to http://www.linfo.org/filters.html
type Filter interface {
	Filter(object *RNode) (*RNode, error)
}

type FilterFunc func(object *RNode) (*RNode, error)

func (f FilterFunc) Filter(object *RNode) (*RNode, error) {
	return f(object)
}

type TypeMeta struct {
	Kind       string
	APIVersion string
}

// ResourceMeta contains the metadata for a both Resource Type and Resource.
type ResourceMeta struct {
	// APIVersion is the apiVersion field of a Resource
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind is the kind field of a Resource
	Kind string `yaml:"kind,omitempty"`
	// ObjectMeta is the metadata field of a Resource
	ObjectMeta `yaml:"metadata,omitempty"`
}

// ObjectMeta contains metadata about a Resource
type ObjectMeta struct {
	// Name is the metadata.name field of a Resource
	Name string `yaml:"name,omitempty"`
	// Namespace is the metadata.namespace field of a Resource
	Namespace string `yaml:"namespace,omitempty"`
	// Labels is the metadata.labels field of a Resource
	Labels map[string]string `yaml:"labels,omitempty"`
	// Annotations is the metadata.annotations field of a Resource.
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// GetIdentifier returns a ResourceIdentifier that includes
// the information needed to uniquely identify a resource in a cluster.
func (m *ResourceMeta) GetIdentifier() ResourceIdentifier {
	return ResourceIdentifier{
		Name:       m.Name,
		Namespace:  m.Namespace,
		APIVersion: m.APIVersion,
		Kind:       m.Kind,
	}
}

var ErrMissingMetadata = fmt.Errorf("missing Resource metadata")

const (
	Trim = "Trim"
	Flow = "Flow"
)

// String returns a string value for a Node, applying the supplied formatting options
func String(node *yaml.Node, opts ...string) (string, error) {
	if node == nil {
		return "", nil
	}
	optsSet := sets.String{}
	optsSet.Insert(opts...)
	if optsSet.Has(Flow) {
		oldStyle := node.Style
		defer func() {
			node.Style = oldStyle
		}()
		node.Style = yaml.FlowStyle
	}

	b := &bytes.Buffer{}
	e := NewEncoder(b)
	err := e.Encode(node)
	e.Close()
	val := b.String()
	if optsSet.Has(Trim) {
		val = strings.TrimSpace(val)
	}
	return val, errors.Wrap(err)
}
