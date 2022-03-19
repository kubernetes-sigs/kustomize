// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/labels"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Add the given labels to the given field specifications.
type plugin struct {
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	FieldSpecs        []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	FieldSpecStrategy string            `json:"fieldSpecStrategy,omitempty" yaml:"fieldSpecStrategy,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	kp := p
	kp.Labels = nil
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
		kp.FieldSpecs, err = types.FsSlice(kp.FieldSpecs).MergeAll(fsSliceAsMap["commonlabels"])
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
	if len(p.Labels) == 0 {
		return nil
	}
	return m.ApplyFilter(labels.Filter{
		Labels:  p.Labels,
		FsSlice: p.FieldSpecs,
	})
}
