// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	h             *resmap.PluginHelpers
	loadedPatches []*resource.Resource
	Paths         []types.PatchStrategicMerge `json:"paths,omitempty" yaml:"paths,omitempty"`
	Patches       string                      `json:"patches,omitempty" yaml:"patches,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	h *resmap.PluginHelpers, c []byte) (err error) {
	p.h = h
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	if len(p.Paths) == 0 && p.Patches == "" {
		return fmt.Errorf("empty file path and empty patch content")
	}
	if len(p.Paths) != 0 {
		for _, onePath := range p.Paths {
			res, err := p.h.ResmapFactory().RF().SliceFromBytes([]byte(onePath))
			if err == nil {
				p.loadedPatches = append(p.loadedPatches, res...)
				continue
			}
			res, err = p.h.ResmapFactory().RF().SliceFromPatches(
				p.h.Loader(), []types.PatchStrategicMerge{onePath})
			if err != nil {
				return err
			}
			p.loadedPatches = append(p.loadedPatches, res...)
		}
	}
	if p.Patches != "" {
		res, err := p.h.ResmapFactory().RF().SliceFromBytes([]byte(p.Patches))
		if err != nil {
			return err
		}
		p.loadedPatches = append(p.loadedPatches, res...)
	}

	if len(p.loadedPatches) == 0 {
		return fmt.Errorf(
			"patch appears to be empty; files=%v, Patch=%s", p.Paths, p.Patches)
	}
	return err
}

func (p *plugin) Transform(m resmap.ResMap) error {
	patches, err := p.h.ResmapFactory().Merge(p.loadedPatches)
	if err != nil {
		return err
	}
	for _, patch := range patches.Resources() {
		target, err := m.GetById(patch.OrgId())
		if err != nil {
			return err
		}
		if err = m.ApplySmPatch(
			resource.MakeIdSet([]*resource.Resource{target}), patch); err != nil {
			return err
		}
	}
	return nil
}
