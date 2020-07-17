// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/filesys"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty/internal/provider"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provenance"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
)

// Kustomizer performs kustomizations.  It's meant to behave
// similarly to the kustomize CLI, and can be used instead of
// performing an exec to a kustomize CLI subprocess.
// To use, load a filesystem with kustomization files (any
// number of overlays and bases), then make a Kustomizer
// injected with the given fileystem, then call Run.
type Kustomizer struct {
	fSys        filesys.FileSystem
	options     *Options
	depProvider *provider.DepProvider
}

// MakeKustomizer returns an instance of Kustomizer.
func MakeKustomizer(fSys filesys.FileSystem, o *Options) *Kustomizer {
	return &Kustomizer{
		fSys:        fSys,
		options:     o,
		depProvider: provider.NewDepProvider(o.UseKyaml),
	}
}

// Run performs a kustomization.
//
// It uses its internal filesystem reference to read the file at
// the given path argument, interpret it as a kustomization.yaml
// file, perform the kustomization it represents, and return the
// resulting resources.
//
// Any files referenced by the kustomization must be present in the
// internal filesystem.  One may call Run any number of times,
// on any number of internal paths (e.g. the filesystem may contain
// multiple overlays, and Run can be called on each of them).
func (b *Kustomizer) Run(path string) (resmap.ResMap, error) {
	resmapFactory := resmap.NewFactory(
		b.depProvider.GetResourceFactory(),
		b.depProvider.GetMerginator())
	lr := fLdr.RestrictionNone
	if b.options.LoadRestrictions == types.LoadRestrictionsRootOnly {
		lr = fLdr.RestrictionRootOnly
	}
	ldr, err := fLdr.NewLoader(lr, path, b.fSys)
	if err != nil {
		return nil, err
	}
	defer ldr.Cleanup()
	kt := target.NewKustTarget(
		ldr,
		b.depProvider.GetFieldValidator(),
		resmapFactory,
		pLdr.NewLoader(b.options.PluginConfig, resmapFactory),
	)
	err = kt.Load()
	if err != nil {
		return nil, err
	}
	var m resmap.ResMap
	m, err = kt.MakeCustomizedResMap()
	if err != nil {
		return nil, err
	}
	if b.options.DoLegacyResourceSort {
		builtins.NewLegacyOrderTransformerPlugin().Transform(m)
	}
	if b.options.AddManagedbyLabel {
		t := builtins.LabelTransformerPlugin{
			Labels: map[string]string{
				konfig.ManagedbyLabelKey: fmt.Sprintf(
					"kustomize-%s", provenance.GetProvenance().Version)},
			FieldSpecs: []types.FieldSpec{{
				Path:               "metadata/labels",
				CreateIfNotPresent: true,
			}},
		}
		t.Transform(m)
	}
	return m, nil
}
