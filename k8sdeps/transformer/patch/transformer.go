// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/transformers"
)

// transformer applies strategic merge patches.
type transformer struct {
	patches []*resource.Resource
	rf      *resource.Factory
}

var _ transformers.Transformer = &transformer{}

// NewTransformer constructs a strategic merge patch transformer.
func NewTransformer(
	slice []*resource.Resource, rf *resource.Factory) (transformers.Transformer, error) {
	if len(slice) == 0 {
		return transformers.NewNoOpTransformer(), nil
	}
	return &transformer{patches: slice, rf: rf}, nil
}

// Transform apply the patches on top of the base resources.
// nolint:ineffassign
func (tf *transformer) Transform(m resmap.ResMap) error {
	patches, err := MergePatches(tf.patches, tf.rf)
	if err != nil {
		return err
	}
	for _, patch := range patches.Resources() {
		target, err := m.GetById(patch.OrgId())
		if err != nil {
			return err
		}
		err = target.Patch(patch.Kunstructured)
		if err != nil {
			return err
		}
	}
	return nil
}
