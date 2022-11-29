// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Find matching image declarations and replace
// the name, tag and/or digest.
type plugin struct {
	ImageTag   types.Image       `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.ImageTag = types.Image{}
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if err := m.ApplyFilter(imagetag.LegacyFilter{
		ImageTag: p.ImageTag,
	}); err != nil {
		return err
	}
	return m.ApplyFilter(imagetag.Filter{
		ImageTag: p.ImageTag,
		FsSlice:  p.FieldSpecs,
	})
}
