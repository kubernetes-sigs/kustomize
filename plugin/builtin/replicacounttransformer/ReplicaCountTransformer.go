// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/replicacount"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/yaml"
)

// Find matching replicas declarations and replace the count.
// Eases the kustomization configuration of replica changes.
type plugin struct {
	Replica           types.Replica     `json:"replica,omitempty" yaml:"replica,omitempty"`
	FieldSpecs        []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	FieldSpecStrategy string            `json:"fieldSpecStrategy,omitempty" yaml:"fieldSpecStrategy,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	kp := p
	kp.Replica = types.Replica{}
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
		kp.FieldSpecs, err = types.FsSlice(kp.FieldSpecs).MergeAll(fsSliceAsMap["replicas"])
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
	found := false
	for _, fs := range p.FieldSpecs {
		matcher := p.createMatcher(fs)
		resList := m.GetMatchingResourcesByAnyId(matcher)
		if len(resList) > 0 {
			found = true
			for _, r := range resList {
				// There are redundant checks in the filter
				// that we'll live with until resolution of
				// https://github.com/kubernetes-sigs/kustomize/issues/2506
				err := r.ApplyFilter(replicacount.Filter{
					Replica:   p.Replica,
					FieldSpec: fs,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	if !found {
		gvks := make([]string, len(p.FieldSpecs))
		for i, replicaSpec := range p.FieldSpecs {
			gvks[i] = replicaSpec.Gvk.String()
		}
		return fmt.Errorf("resource with name %s does not match a config with the following GVK %v",
			p.Replica.Name, gvks)
	}

	return nil
}

// Match Replica.Name and FieldSpec
func (p *plugin) createMatcher(fs types.FieldSpec) resmap.IdMatcher {
	return func(r resid.ResId) bool {
		return r.Name == p.Replica.Name && r.Gvk.IsSelected(&fs.Gvk)
	}
}
