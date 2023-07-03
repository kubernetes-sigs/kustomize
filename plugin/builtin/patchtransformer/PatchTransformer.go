// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"sigs.k8s.io/kustomize/api/filters/patchjson6902"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	smPatches []*resource.Resource // strategic-merge patches
	jsonPatch jsonpatch.Patch      // json6902 patch
	Path      string               `json:"path,omitempty"    yaml:"path,omitempty"`
	Patch     string               `json:"patch,omitempty"   yaml:"patch,omitempty"`
	Target    *types.Selector      `json:"target,omitempty"  yaml:"target,omitempty"`
	Options   map[string]bool      `json:"options,omitempty" yaml:"options,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
	h *resmap.PluginHelpers, c []byte) error {
	err := yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	p.Patch = strings.TrimSpace(p.Patch)
	if p.Patch == "" && p.Path == "" {
		return fmt.Errorf(
			"must specify one of patch and path in\n%s", string(c))
	}
	if p.Patch != "" && p.Path != "" {
		return fmt.Errorf(
			"patch and path can't be set at the same time\n%s", string(c))
	}
	if p.Path != "" {
		loaded, loadErr := h.Loader().Load(p.Path)
		if loadErr != nil {
			return loadErr
		}
		p.Patch = string(loaded)
	}

	patchesSM, errSM := h.ResmapFactory().RF().SliceFromBytes([]byte(p.Patch))
	patchJson, errJson := jsonPatchFromBytes([]byte(p.Patch))
	if (errSM == nil && errJson == nil) ||
		(patchesSM != nil && patchJson != nil) {
		return fmt.Errorf(
			"illegally qualifies as both an SM and JSON patch: [%v]",
			p.Patch)
	}
	if errSM != nil && errJson != nil {
		return fmt.Errorf(
			"unable to parse SM or JSON patch from [%v]", p.Patch)
	}
	if errSM == nil {
		p.smPatches = patchesSM
		for _, loadedPatch := range p.smPatches {
			if p.Options["allowNameChange"] {
				loadedPatch.AllowNameChange()
			}
			if p.Options["allowKindChange"] {
				loadedPatch.AllowKindChange()
			}
		}
	} else {
		p.jsonPatch = patchJson
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if p.smPatches == nil {
		return p.transformJson6902(m, p.jsonPatch)
	}
	// The patch was a strategic merge patch
	return p.transformStrategicMerge(m)
}

// transformStrategicMerge applies each loaded strategic merge patch
// to the resource in the ResMap that matches the identifier of the patch.
// If only one patch is specified, the Target can be used instead.
func (p *plugin) transformStrategicMerge(m resmap.ResMap) error {
	for _, patch := range p.smPatches {
		if p.Target == nil {
			target, err := m.GetById(patch.OrgId())
			if err != nil {
				return fmt.Errorf("unable to find patch target %s: %w", patch.OrgId().Name, err)
			}
			if err := target.ApplySmPatch(patch); err != nil {
				return fmt.Errorf("%w", err)
			}
			continue
		}

		selected, err := m.Select(*p.Target)
		if err != nil {
			return fmt.Errorf("unable to find patch target in ResMap %s: %w", p.Target, err)
		}
		if err := m.ApplySmPatch(resource.MakeIdSet(selected), patch); err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}

// transformJson6902 applies the provided json6902 patch
// to all the resources in the ResMap that match the Target.
func (p *plugin) transformJson6902(m resmap.ResMap, patch jsonpatch.Patch) error {
	if p.Target == nil {
		return fmt.Errorf("must specify a target for patch %s", p.Patch)
	}
	resources, err := m.Select(*p.Target)
	if err != nil {
		return err
	}
	for _, res := range resources {
		res.StorePreviousId()
		internalAnnotations := kioutil.GetInternalAnnotations(&res.RNode)
		err = res.ApplyFilter(patchjson6902.Filter{
			Patch: p.Patch,
		})
		if err != nil {
			return err
		}

		annotations := res.GetAnnotations()
		for key, value := range internalAnnotations {
			annotations[key] = value
		}
		err = res.SetAnnotations(annotations)
	}
	return nil
}

// jsonPatchFromBytes loads a Json 6902 patch from
// a bytes input
func jsonPatchFromBytes(
	in []byte) (jsonpatch.Patch, error) {
	ops := string(in)
	if ops == "" {
		return nil, fmt.Errorf("empty json patch operations")
	}

	if ops[0] != '[' {
		jsonOps, err := yaml.YAMLToJSON(in)
		if err != nil {
			return nil, err
		}
		ops = string(jsonOps)
	}
	return jsonpatch.DecodePatch([]byte(ops))
}
