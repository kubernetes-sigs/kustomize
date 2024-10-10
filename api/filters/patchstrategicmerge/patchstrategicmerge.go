// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patchstrategicmerge

import (
	"cmp"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

type Filter struct {
	Patch        *yaml.RNode
	MergeOptions *yaml.MergeOptions
}

var _ kio.Filter = Filter{}

// Filter does a strategic merge patch, which can delete nodes.
func (pf Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	var result []*yaml.RNode
	mergeOptions := *cmp.Or(pf.MergeOptions, &yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListPrepend,
	})
	for i := range nodes {
		r, err := merge2.Merge(
			pf.Patch, nodes[i],
			mergeOptions,
		)
		if err != nil {
			return nil, err
		}
		if r != nil {
			result = append(result, r)
		}
	}
	return result, nil
}
