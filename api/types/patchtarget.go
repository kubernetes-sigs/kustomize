// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"sigs.k8s.io/kustomize/api/resid"
)

// PatchTarget represents the kubernetes object that the patch is applied to
type PatchTarget struct {
	resid.Gvk `json:",inline,omitempty" yaml:",inline,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string `json:"name" yaml:"name"`
}

// ToSelector converts a PatchTarget to a Selector.
func (target *PatchTarget) ToSelector() Selector {
	return Selector{Name: target.Name, Namespace: target.Namespace, Gvk: target.Gvk}
}
