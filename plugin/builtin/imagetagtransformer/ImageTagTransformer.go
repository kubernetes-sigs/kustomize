// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Find matching image declarations and replace
// the name, tag and/or digest.
type plugin struct {
	ImageTag          types.Image       `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	FieldSpecs        []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	FieldSpecStrategy string            `json:"fieldSpecStrategy,omitempty" yaml:"fieldSpecStrategy,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	kp := p
	kp.ImageTag = types.Image{}
	kp.FieldSpecs = nil
	if err = yaml.Unmarshal(c, kp); err != nil {
		return err
	}
	p = kp
	switch kp.FieldSpecStrategy {
	case types.FieldSpecStrategy[types.Merge]:
		fsSliceAsMap, err := builtinpluginconsts.GetFsSliceAsMap()
		if err != nil {
			return err
		}
		kp.FieldSpecs, err = types.FsSlice(kp.FieldSpecs).MergeAll(fsSliceAsMap["images"])
		if err != nil {
			return err
		}
		p.FieldSpecs = kp.FieldSpecs
	case types.FieldSpecStrategy[types.Replace]:
		p.FieldSpecs = kp.FieldSpecs
	}
	return nil
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
