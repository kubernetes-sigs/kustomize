// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/labels"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Add the given labels to the given field specifications.
type plugin struct {
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Labels = nil
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if len(p.Labels) == 0 {
		return nil
	}
	return m.ApplyFilter(labels.Filter{
		Labels:  p.Labels,
		FsSlice: p.FieldSpecs,
	})
}
