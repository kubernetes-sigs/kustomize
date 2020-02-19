// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// visitor is implemented by structs which need to walk the configuration.
// visitor is provided to accept to walk configuration
type visitor interface {
	// visitScalar is called for each scalar field value on a resource
	// node is the scalar field value
	// path is the path to the field; path elements are separated by '.'
	visitScalar(node *yaml.RNode, path string) error
}

// accept invokes the appropriate function on v for each field in object
func accept(v visitor, object *yaml.RNode) error {
	return acceptImpl(v, object, "")
}

// acceptImpl implements accept using recursion
func acceptImpl(v visitor, object *yaml.RNode, p string) error {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		// Traverse the child of the document
		return accept(v, yaml.NewRNode(object.YNode()))
	case yaml.MappingNode:
		return object.VisitFields(func(node *yaml.MapNode) error {
			// Traverse each field value
			return acceptImpl(v, node.Value, p+"."+node.Key.YNode().Value)
		})
	case yaml.SequenceNode:
		return object.VisitElements(func(node *yaml.RNode) error {
			// Traverse each list element
			return acceptImpl(v, node, p)
		})
	case yaml.ScalarNode:
		// Visit the scalar field
		return v.visitScalar(object, p)
	}
	return nil
}
