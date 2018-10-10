/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package transformer provides transformer factory
package transformer

import (
	"sigs.k8s.io/kustomize/internal/k8sdeps/transformer/hash"
	"sigs.k8s.io/kustomize/internal/k8sdeps/transformer/patch"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

// factoryImpl makes patch transformer and name hash transformer
type factoryImpl struct{}

// NewFactoryImpl makes a new factoryImpl instance
func NewFactoryImpl() *factoryImpl {
	return &factoryImpl{}
}

// MakePatchTransformer makes a new patch transformer
func (p *factoryImpl) MakePatchTransformer(slice []*resource.Resource, rf *resource.Factory) (transformers.Transformer, error) {
	return patch.NewPatchTransformer(slice, rf)
}

// MakeHashTransformer makes a new name hash transformer
func (p *factoryImpl) MakeHashTransformer() transformers.Transformer {
	return hash.NewNameHashTransformer()
}
