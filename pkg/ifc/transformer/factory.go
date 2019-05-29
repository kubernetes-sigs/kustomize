/// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package patch holds miscellaneous interfaces used by kustomize.
package transformer

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

// Factory makes transformers
type Factory interface {
	MakePatchTransformer(slice []*resource.Resource, rf *resource.Factory) (transformers.Transformer, error)
	MakeHashTransformer() transformers.Transformer
	MakeInventoryTransformer(
		p *types.Inventory,
		ldr ifc.Loader,
		namespace string,
		gp types.GarbagePolicy) transformers.Transformer
}
