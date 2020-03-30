// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package imagetag

import (
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	// imageTag is the tag we want to apply to the inputs
	ImageTag types.Image `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`

	// FsSlice contains the FieldSpecs to locate the namespace field
	FsSlice types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

var _ kio.Filter = Filter{}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

func (f Filter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	if err := node.PipeE(fsslice.Filter{
		FsSlice:  f.FsSlice,
		SetValue: updateImageTagFn(f.ImageTag),
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func updateImageTagFn(imageTag types.Image) fsslice.SetFn {
	return func(node *yaml.RNode) error {
		return node.PipeE(imageTagUpdater{
			ImageTag: imageTag,
		})
	}
}
