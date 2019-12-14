// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by pluginator on HashTransformer; DO NOT EDIT.
// pluginator {unknown  1970-01-01T00:00:00Z  }

package builtins

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
)

type HashTransformerPlugin struct {
	hasher ifc.KunstructuredHasher
}

func (p *HashTransformerPlugin) Config(
	h *resmap.PluginHelpers, _ []byte) (err error) {
	p.hasher = h.ResmapFactory().RF().Hasher()
	return nil
}

// Transform appends hash to generated resources.
func (p *HashTransformerPlugin) Transform(m resmap.ResMap) error {
	for _, res := range m.Resources() {
		if res.NeedHashSuffix() {
			h, err := p.hasher.Hash(res)
			if err != nil {
				return err
			}
			res.SetName(fmt.Sprintf("%s-%s", res.GetName(), h))
		}
	}
	return nil
}

func NewHashTransformerPlugin() resmap.TransformerPlugin {
	return &HashTransformerPlugin{}
}
