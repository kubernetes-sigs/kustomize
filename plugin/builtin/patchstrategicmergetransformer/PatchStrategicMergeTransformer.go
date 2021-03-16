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
	loadedPatches []*resource.Resource
	Paths         []types.PatchStrategicMerge `json:"paths,omitempty" yaml:"paths,omitempty"`
	Patches       string                      `json:"patches,omitempty" yaml:"patches,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	h *resmap.PluginHelpers, c []byte) (err error) {
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	if len(p.Paths) == 0 && p.Patches == "" {
		return fmt.Errorf("empty file path and empty patch content")
	}
	if len(p.Paths) != 0 {
		patches, err := loadFromPaths(h, p.Paths)
		if err != nil {
			return err
		}
		p.loadedPatches = append(p.loadedPatches, patches...)
	}
	if p.Patches != "" {
		patches, err := h.ResmapFactory().RF().SliceFromBytes([]byte(p.Patches))
		if err != nil {
			return err
		}
		p.loadedPatches = append(p.loadedPatches, patches...)
	}
	if len(p.loadedPatches) == 0 {
		return fmt.Errorf(
			"patch appears to be empty; files=%v, Patch=%s", p.Paths, p.Patches)
	}
	// TODO(#3723): Delete conflict detection.
	// Since #1500 closed, the conflict detector in use doesn't do
	// anything useful.  The resmap returned by this method hasn't
	// been used for many releases.  Leaving code as a comment to
	// aid in deletion (fixing #3723).
	// _, err = h.ResmapFactory().ConflatePatches(p.loadedPatches)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func loadFromPaths(
	h *resmap.PluginHelpers,
	paths []types.PatchStrategicMerge) (
	result []*resource.Resource, err error) {
	var patches []*resource.Resource
	for _, path := range paths {
		// For legacy reasons, attempt to treat the path string as
		// actual patch content.
		patches, err = h.ResmapFactory().RF().SliceFromBytes([]byte(path))
		if err != nil {
			// Failing that, treat it as a file path.
			patches, err = h.ResmapFactory().RF().SliceFromPatches(
				h.Loader(), []types.PatchStrategicMerge{path})
			if err != nil {
				return
			}
		}
		result = append(result, patches...)
	}
	return
}

func (p *plugin) Transform(m resmap.ResMap) error {
	for _, patch := range p.loadedPatches {
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
