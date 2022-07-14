// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package localizer contains utilities for the command kustomize localize, which is
// documented under proposals/localize-command or at
// https://github.com/kubernetes-sigs/kustomize/blob/master/proposals/22-04-localize-command.md
package localizer

import (
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	plgnLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	kustTg "sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

type localizer struct {
	fSys      filesys.FileSystem
	ll        *locLoader
	vldtr     ifc.Validator
	rmFactory *resmap.Factory
	plgnLdr   *plgnLdr.Loader
}

// validation happens in cmd/localize.go
func Run(fSys filesys.FileSystem, targetArg string, scope string, newDir string) error {
	ll, err := newLocLoader(targetArg, scope, newDir, fSys)
	if err != nil {
		return err
	}
	defer ll.ldr.Cleanup()

	// TODO: understand this better
	depProvider := provider.NewDepProvider()
	rmFactory := resmap.NewFactory(depProvider.GetResourceFactory())

	err = (&localizer{
		fSys:      fSys,
		ll:        ll,
		vldtr:     depProvider.GetFieldValidator(),
		rmFactory: rmFactory,
		plgnLdr:   plgnLdr.NewLoader(krusty.MakeDefaultOptions().PluginConfig, rmFactory, filesys.MakeFsOnDisk()),
	}).localize()
	return err
}

func (lcr *localizer) localizeFile(path string) (string, error) {
	content, dstPath, err := lcr.ll.load(path)
	if err != nil {
		return "", err
	}

	file := lcr.ll.dest().Join(dstPath)
	err = lcr.fSys.MkdirAll(filepath.Dir(file))
	if err != nil {
		return "", err
	}
	err = lcr.fSys.WriteFile(file, content)
	if err != nil {
		return "", err
	}

	return dstPath, nil
}

func (lcr *localizer) localize() error {
	kt := kustTg.NewKustTarget(lcr.ll.ldr, lcr.vldtr, lcr.rmFactory, lcr.plgnLdr)
	err := kt.Load()
	if err != nil {
		return err
	}
	kust := kt.Kustomization()

	localizedResources := make([]string, len(kust.Resources))
	for i, path := range kust.Resources {
		var localizedPath string
		if localizedPath, err = lcr.localizeDir(path); err != nil {
			if localizedPath, err = lcr.localizeFile(path); err != nil {
				return err
			}
		}
		localizedResources[i] = localizedPath
	}
	kust.Resources = localizedResources

	// TODO: understand this better
	content, err := yaml.Marshal(kust)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to serialize kustomization file at '%s'", lcr.ll.ldr.Root())
	}
	return lcr.fSys.WriteFile(lcr.ll.dest().Join(konfig.DefaultKustomizationFileName()), content)
}

func (lcr *localizer) localizeDir(path string) (string, error) {
	ll, dstPath, err := lcr.ll.new(path)
	if err != nil {
		return "", err
	}
	defer ll.ldr.Cleanup()

	err = (&localizer{
		fSys:      lcr.fSys,
		ll:        ll,
		vldtr:     lcr.vldtr,
		rmFactory: lcr.rmFactory,
		plgnLdr:   lcr.plgnLdr,
	}).localize()
	if err != nil {
		return "", err
	}

	return dstPath, nil
}
