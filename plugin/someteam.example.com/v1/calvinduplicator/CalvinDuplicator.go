// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

// A silly plugin that duplicates resources with a given name.
// It looks for resources named $Name, and duplicates them N times,
// creating resources named ${Name}-1, ..., ${Name}-N.
// See https://calvinandhobbes.fandom.com/wiki/Duplicator
type plugin struct {
	// Name of resource to duplicate.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Count is how many duplicates to make.
	Count int `json:"count,omitempty" yaml:"count,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

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
				err := c.SetName(fmt.Sprintf("%s-%d", p.Name, i))
				if err != nil {
					return err
				}
				err = m.Append(c)
				if err != nil {
					return err
				}
			}
		} else {
			err := m.Append(r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
