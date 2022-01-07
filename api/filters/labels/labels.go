// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package labels

import (
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type labelMap map[string]string

// Filter sets labels.
type Filter struct {
	// Labels is the set of labels to apply to the inputs
	Labels labelMap `yaml:"labels,omitempty"`

	// FsSlice identifies the label fields.
	FsSlice types.FsSlice

	// SetEntryCallback is invoked each time a label is applied
	// Example use cases:
	//   - Tracking all paths where labels have been applied
	SetEntryCallback func(key, value, tag string, node *yaml.RNode)
}

var _ kio.Filter = Filter{}

func (f Filter) setEntry(key, value, tag string) filtersutil.SetFn {
	baseSetEntryFunc := filtersutil.SetEntry(key, value, tag)
	return func(node *yaml.RNode) error {
		if f.SetEntryCallback != nil {
			f.SetEntryCallback(key, value, tag, node)
		}
		return baseSetEntryFunc(node)
	}
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	keys := yaml.SortedMapKeys(f.Labels)
	_, err := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			for _, k := range keys {
				if err := node.PipeE(fsslice.Filter{
					FsSlice: f.FsSlice,
					SetValue: f.setEntry(
						k, f.Labels[k], yaml.NodeTagString),
					CreateKind: yaml.MappingNode, // Labels are MappingNodes.
					CreateTag:  yaml.NodeTagMap,
				}); err != nil {
					return nil, err
				}
			}
			return node, nil
		})).Filter(nodes)
	return nodes, err
}
