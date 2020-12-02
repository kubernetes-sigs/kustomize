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
// TODO: Detect conflicts, and return an error.
// https://github.com/kubernetes-sigs/kustomize/issues/3303
func (m Merginator) Merge(
	resources []*resource.Resource) (resmap.ResMap, error) {
	rm := resmap.New()
	for i := range resources {
		rm.Append(resources[i])
	}
	return rm, nil
}
