// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package prefixsuffix

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter applies resource name prefix's and suffix's using the fieldSpecs
type Filter struct {
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty" yaml:"suffix,omitempty"`

	FsSlice types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

var _ kio.Filter = Filter{}

func (ns Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(ns.run)).Filter(nodes)
}

// Run runs the filter on a single node rather than a slice
func (ns Filter) run(node *yaml.RNode) (*yaml.RNode, error) {
	// transformations based on data -- :)
	err := node.PipeE(fsslice.Filter{
		FsSlice:    ns.FsSlice,
		SetValue:   ns.set,
		CreateKind: yaml.ScalarNode, // Name is a ScalarNode
		CreateTag:  yaml.StringTag,
	})
	return node, err
}

func (ns Filter) set(node *yaml.RNode) error {
	return fsslice.SetScalar(fmt.Sprintf(
		"%s%s%s", ns.Prefix, node.YNode().Value, ns.Suffix))(node)
}
