// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge

import (
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

// Merginator implements resmap.Merginator using kyaml libs.
type Merginator struct {
}

var _ resmap.Merginator = (*Merginator)(nil)

func NewMerginator(_ *resource.Factory) *Merginator {
	return &Merginator{}
}

// Merge implements resmap.Merginator
func (m Merginator) Merge(
	resources []*resource.Resource) (resmap.ResMap, error) {
	panic("TODO(#Merginator): implement Merge")
}
