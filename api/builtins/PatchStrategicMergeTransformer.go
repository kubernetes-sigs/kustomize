// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by pluginator on PatchStrategicMergeTransformer; DO NOT EDIT.
// pluginator {unknown  1970-01-01T00:00:00Z  }

package builtins

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type PatchStrategicMergeTransformerPlugin struct {
	h             *resmap.PluginHelpers
	loadedPatches []*resource.Resource
	Paths         []types.PatchStrategicMerge `json:"paths,omitempty" yaml:"paths,omitempty"`
	Patches       string                      `json:"patches,omitempty" yaml:"patches,omitempty"`
}

func (p *PatchStrategicMergeTransformerPlugin) Config(
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

func (p *PatchStrategicMergeTransformerPlugin) Transform(m resmap.ResMap) error {
	patches, err := p.h.ResmapFactory().MergePatches(p.loadedPatches)
	if err != nil {
		return err
	}
	for _, patch := range patches.Resources() {
		target, err := m.GetById(patch.OrgId())
		if err != nil {
			return err
		}
		err = target.Patch(patch.Kunstructured)
		if err != nil {
			return err
		}
		// remove the resource from resmap
		// when the patch is to $patch: delete that target
		if len(target.Map()) == 0 {
			err = m.Remove(target.CurId())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func NewPatchStrategicMergeTransformerPlugin() resmap.TransformerPlugin {
	return &PatchStrategicMergeTransformerPlugin{}
}
