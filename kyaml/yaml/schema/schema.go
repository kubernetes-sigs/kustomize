// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package schema contains libraries for working with the yaml and openapi packages.
package schema

import (
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// IsAssociative returns true if all elements in the list contain an AssociativeSequenceKey
// as a field.
func IsAssociative(schema *openapi.ResourceSchema, nodes []*yaml.RNode, infer bool) bool {
	if schema != nil {
		// use the schema to identify if the list is associative
		s, _ := schema.PatchStrategyAndKey()
		return s == "merge"
	}
	if !infer {
		return false
	}

	for i := range nodes {
		node := nodes[i]
		if yaml.IsEmpty(node) {
			continue
		}
		if node.IsAssociative() {
			return true
		}
	}
	return false
}
