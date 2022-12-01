// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
)

type plugin struct {
	hasher ifc.KustHasher
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
	h *resmap.PluginHelpers, _ []byte) (err error) {
	p.hasher = h.ResmapFactory().RF().Hasher()
	return nil
}

// Transform appends hash to generated resources.
func (p *plugin) Transform(m resmap.ResMap) error {
	for _, res := range m.Resources() {
		if res.NeedHashSuffix() {
			h, err := res.Hash(p.hasher)
			if err != nil {
				return err
			}
			res.StorePreviousId()
			res.SetName(fmt.Sprintf("%s-%s", res.GetName(), h))
		}
	}
	return nil
}
