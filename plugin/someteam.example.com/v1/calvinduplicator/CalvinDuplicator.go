// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

const calvinName = "calvin"

// Look for resources named $Name, and duplicate them N times,
// leaving resources named ${Name}-1, ..., ${Name}-N.
type plugin struct {
	// Name of resource to duplicate.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Count is how many duplicates to make.
	Count int `json:"count,omitempty" yaml:"count,omitempty"`
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(_ *resmap.PluginHelpers, c []byte) error {
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	list := m.Resources()
	m.Clear()
	for _, r := range list {
		if r.GetName() == p.Name {
			for i := 1; i <= p.Count; i++ {
				c := r.DeepCopy()
				c.SetName(fmt.Sprintf("%s-%d", p.Name, i))
				m.Append(c)
			}
		} else {
			m.Append(r)
		}
	}
	return nil
}
