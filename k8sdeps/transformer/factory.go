// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package transformer provides transformer factory
package transformer

import (
	"sigs.k8s.io/kustomize/k8sdeps/transformer/patch"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

// FactoryImpl makes patch transformer and name hash transformer
type FactoryImpl struct{}

// NewFactoryImpl makes a new factoryImpl instance
func NewFactoryImpl() *FactoryImpl {
	return &FactoryImpl{}
}

// MakePatchTransformer makes a new patch transformer
func (p *FactoryImpl) MakePatchTransformer(
	slice []*resource.Resource,
	rf *resource.Factory) (transformers.Transformer, error) {
	return patch.NewTransformer(slice, rf)
}
