// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package annotations

import (
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	// Annotations is the set of annotations to apply to the inputs
	Annotations map[string]string `yaml:"annotations,omitempty"`

	// FsSlice contains the FieldSpecs to locate the namespace field
	FsSlice types.FsSlice
}

var _ kio.Filter = Filter{}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range nodes {
		if err := f.run(nodes[i]); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// run applies the filter to a single node.
func (f Filter) run(node *yaml.RNode) error {
	for key, value := range f.Annotations {
		if err := node.PipeE(fsslice.Filter{
			FsSlice:    f.FsSlice,
			SetValue:   fsslice.SetEntry(key, value),
			CreateKind: yaml.MappingNode, // Annotations are MappingNodes.
		}); err != nil {
			return err
		}
	}
	return nil
}
