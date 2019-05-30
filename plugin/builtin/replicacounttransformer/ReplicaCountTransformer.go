// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run sigs.k8s.io/kustomize/plugin/pluginator
package main

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/replica"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/yaml"
)

// Find matching replicas declarations and replace the count.
// Eases the kustomization configuration of replica changes.
type plugin struct {
	Replica    replica.Replica    `json:"replica,omitempty" yaml:"replica,omitempty"`
	FieldSpecs []config.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {

	p.Replica = replica.Replica{}
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	matcher := func(r resid.ResId) bool {
		return r.ItemId.Name == p.Replica.Name
	}

	for _, r := range m.GetMatchingIds(matcher) {
		kMap := m[r].Kunstructured.Map()

		specInterface, ok := kMap["spec"]
		if !ok {
			return errors.New("'spec' not specified, replicas cannot be modified")
		}

		if spec, ok := specInterface.(map[string]interface{}); ok {
			spec["replicas"] = p.Replica.Count
			kMap["spec"] = spec
		} else {
			return errors.New("'spec' not structured as expected")
		}

		m[r].Kunstructured.SetMap(kMap)
	}

	return nil
}
