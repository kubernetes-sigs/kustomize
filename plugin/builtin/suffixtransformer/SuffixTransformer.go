// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/suffix"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Add the given suffix to the field
type plugin struct {
	Suffix            string        `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	FieldSpecs        types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	FieldSpecStrategy string        `json:"fieldSpecStrategy,omitempty" yaml:"fieldSpecStrategy,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// TODO: Make this gvk skip list part of the config.
var suffixFieldSpecsToSkip = types.FsSlice{
	{Gvk: resid.Gvk{Kind: "CustomResourceDefinition"}},
	{Gvk: resid.Gvk{Group: "apiregistration.k8s.io", Kind: "APIService"}},
	{Gvk: resid.Gvk{Kind: "Namespace"}},
}

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	kp := p
	kp.Suffix = ""
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
		kp.FieldSpecs, err = types.FsSlice(kp.FieldSpecs).MergeAll(fsSliceAsMap["namesuffix"])
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
	// Even if the Suffix is empty we want to proceed with the
	// transformation. This allows to add contextual information
	// to the resources (AddNameSuffix).
	for _, r := range m.Resources() {
		// TODO: move this test into the filter (i.e. make a better filter)
		if p.shouldSkip(r.OrgId()) {
			continue
		}
		id := r.OrgId()
		// current default configuration contains
		// only one entry: "metadata/name" with no GVK
		for _, fs := range p.FieldSpecs {
			// TODO: this is redundant to filter (but needed for now)
			if !id.IsSelected(&fs.Gvk) {
				continue
			}
			// TODO: move this test into the filter.
			if fs.Path == "metadata/name" {
				// "metadata/name" is the only field.
				// this will add a suffix to the resource
				// even if it is empty

				r.AddNameSuffix(p.Suffix)
				if p.Suffix != "" {
					// TODO: There are multiple transformers that can change a resource's name, and each makes a call to
					// StorePreviousID(). We should make it so that we only call StorePreviousID once per kustomization layer
					// to avoid storing intermediate names between transformations, to prevent intermediate name conflicts.
					r.StorePreviousId()
				}
			}
			if err := r.ApplyFilter(suffix.Filter{
				Suffix:    p.Suffix,
				FieldSpec: fs,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *plugin) shouldSkip(id resid.ResId) bool {
	for _, path := range suffixFieldSpecsToSkip {
		if id.IsSelected(&path.Gvk) {
			return true
		}
	}
	return false
}
