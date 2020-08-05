// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kunstruct

import (
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
)

// kustHash computes a hash of an unstructured object.
type kustHash struct{}

// NewKustHash returns a kustHash object
func NewKustHash() *kustHash {
	return &kustHash{}
}

// Hash returns a hash of the given object
func (h *kustHash) Hash(m ifc.Kunstructured) (string, error) {
	node, err := filtersutil.GetRNode(m)
	if err != nil {
		return "", err
	}
	return hasher.HashRNode(node)
}
