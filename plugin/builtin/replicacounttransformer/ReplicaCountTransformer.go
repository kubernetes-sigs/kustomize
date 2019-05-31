// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run sigs.k8s.io/kustomize/plugin/pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
	"sigs.k8s.io/yaml"
)

const (
	fldReplica = "replicas"
	fldSpec    = "spec"
)

// Find matching replicas declarations and replace the count.
// Eases the kustomization configuration of replica changes.
type plugin struct {
	Replica types.Replica `json:"replica,omitempty" yaml:"replica,omitempty"`
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {

	p.Replica = types.Replica{}
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	matcher := func(r resid.ResId) bool {
		return r.ItemId.Name == p.Replica.Name
	}

	for _, r := range m.GetMatchingIds(matcher) {
		kMap := m[r].Map()

		specInterface, ok := kMap[fldSpec]
		if !ok {
			return fmt.Errorf("object %s missing field %s, cannot update %s",
				p.Replica.Name, fldSpec, fldReplica)
		}

		if spec, ok := specInterface.(map[string]interface{}); ok {
			spec[fldReplica] = p.Replica.Count
			kMap[fldSpec] = spec
		} else {
			return fmt.Errorf("object %s has a malformed %s", p.Replica.Name, fldSpec)
		}
	}

	return nil
}
