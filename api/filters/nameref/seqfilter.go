// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package nameref

import (
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type setFn func(*yaml.RNode) error

type seqFilter struct {
	setScalarFn  setFn
	setMappingFn setFn
}

func (sf seqFilter) Filter(node *yaml.RNode) (*yaml.RNode, error) {
	if yaml.IsMissingOrNull(node) {
		return node, nil
	}
	switch node.YNode().Kind {
	case yaml.ScalarNode:
		// Kind: Role/ClusterRole
		// FieldSpec is rules.resourceNames
		err := sf.setScalarFn(node)
		return node, err
	case yaml.MappingNode:
		// Kind: RoleBinding/ClusterRoleBinding
		// FieldSpec is subjects
		// Note: The corresponding fieldSpec had been changed from
		// from path: subjects/name to just path: subjects. This is
		// what get mutatefield to request the mapping of the whole
		// map containing namespace and name instead of just a simple
		// string field containing the name
		err := sf.setMappingFn(node)
		return node, err
	case yaml.AliasNode:
		// YAML spec forbids circular aliases; go-yaml parser guarantees
		// that Alias chains are acyclic, so recursion always terminates.
		if node.YNode().Alias == nil {
			return node, nil
		}
		return sf.Filter(yaml.NewRNode(node.YNode().Alias))
	default:
		// Skip nodes of unexpected kinds rather than failing.
		log.Printf("nameReference: skipping unexpected node kind %d in sequence element"+
			" (if this is unexpected, check your nameReference fieldSpecs)",
			node.YNode().Kind)
		return node, nil
	}
}

// applyFilterToSeq will apply the filter to each element in the sequence node
func applyFilterToSeq(filter yaml.Filter, node *yaml.RNode) error {
	if node.YNode().Kind != yaml.SequenceNode {
		return fmt.Errorf("expect a sequence node but got %v", node.YNode().Kind)
	}

	for _, elem := range node.Content() {
		rnode := yaml.NewRNode(elem)
		err := rnode.PipeE(filter)
		if err != nil {
			return err
		}
	}

	return nil
}
