// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package merge contains libraries for merging Resources and Patches
package filters

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

// GrepFilter merges Resources with the Group/Version/Kind/Namespace/Name together using
// a 2-way merge strategy.
//
// - Fields set to null in the source will be cleared from the destination
// - Fields with matching keys will be merged recursively
// - Lists with an associative key (e.g. name) will have their elements merged using the key
// - List without an associative key will have the dest list replaced by the source list
type MergeFilter struct {
	Reverse bool
}

type mergeKey struct {
	apiVersion string
	kind       string
	namespace  string
	name       string
}

// GrepFilter implements kio.GrepFilter by merge Resources with the same G/V/K/NS/N
func (c MergeFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	// invert the merge precedence
	if c.Reverse {
		for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1 {
			input[i], input[j] = input[j], input[i]
		}
	}

	// index the Resources by G/V/K/NS/N
	index := map[mergeKey][]*yaml.RNode{}
	for i := range input {
		meta, err := input[i].GetMeta()
		if err != nil {
			return nil, err
		}
		key := mergeKey{
			apiVersion: meta.ApiVersion,
			kind:       meta.Kind,
			namespace:  meta.Namespace,
			name:       meta.Name,
		}
		index[key] = append(index[key], input[i])
	}

	// merge each of the G/V/K/NS/N lists
	var output []*yaml.RNode
	var err error
	for k := range index {
		var merged *yaml.RNode
		resources := index[k]
		for i := range resources {
			patch := resources[i]
			if merged == nil {
				// first resources, don't merge it
				merged = resources[i]
			} else {
				merged, err = merge2.Merge(patch, merged)
				if err != nil {
					return nil, err
				}
			}
		}
		output = append(output, merged)
	}
	return output, nil
}
