// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// Localizer encapsulates all state needed to localize the root at ldr.
type Localizer struct {
	fSys filesys.FileSystem

	// kusttarget fields
	validator ifc.Validator
	rFactory  *resmap.Factory
	pLdr      *pLdr.Loader

	// underlying type is Loader
	ldr ifc.Loader

	// destination directory in newDir that mirrors ldr's current root.
	dst filesys.ConfirmedDir
}

// NewLocalizer is the factory method for Localizer
func NewLocalizer(ldr *Loader, validator ifc.Validator, rFactory *resmap.Factory, pLdr *pLdr.Loader) (*Localizer, error) {
	toDst, err := filepath.Rel(ldr.args.Scope.String(), ldr.Root())
	if err != nil {
		log.Fatalf("cannot find path from %q to child directory %q: %s", ldr.args.Scope, ldr.Root(), err)
	}
	dst := ldr.args.NewDir.Join(toDst)
	if err = ldr.fSys.MkdirAll(dst); err != nil {
		return nil, errors.WrapPrefixf(err, "unable to create directory in localize destination")
	}
	return &Localizer{
		fSys:      ldr.fSys,
		validator: validator,
		rFactory:  rFactory,
		pLdr:      pLdr,
		ldr:       ldr,
		dst:       filesys.ConfirmedDir(dst),
	}, nil
}

// Localize localizes the root that lc is at
func (lc *Localizer) Localize() error {
	kt := target.NewKustTarget(lc.ldr, lc.validator, lc.rFactory, lc.pLdr)
	err := kt.Load()
	if err != nil {
		return errors.Wrap(err)
	}
	kust, err := lc.processKust(kt)
	if err != nil {
		return err
	}
	content, err := yaml.Marshal(kust)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to serialize localized kustomization file")
	}
	if err = lc.fSys.WriteFile(lc.dst.Join(konfig.DefaultKustomizationFileName()), content); err != nil {
		return errors.WrapPrefixf(err, "unable to write localized kustomization file")
	}
	return nil
}

// processKust returns a copy of the kustomization at kt with paths localized.
func (lc *Localizer) processKust(kt *target.KustTarget) (*types.Kustomization, error) {
	kust := kt.Kustomization()
	for fieldName, patches := range map[string][]types.Patch{
		"patches":         kust.Patches,
		"patchesJson6902": kust.PatchesJson6902,
	} {
		for i := range patches {
			if patches[i].Path != "" {
				newPath, err := lc.localizeFile(patches[i].Path)
				if err != nil {
					return nil, errors.WrapPrefixf(err, "unable to localize path %q from field %s", patches[i].Path,
						fieldName)
				}
				patches[i].Path = newPath
			}
		}
	}
	// TODO(annasong): localize all other kustomization fields: resources, components, crds, configurations,
	// openapi, patchesStrategicMerge, replacements, configMapGenerators, secretGenerators
	// TODO(annasong): localize built-in plugins under generators, transformers, and validators fields
	return &kust, nil
}

// localizeFile localizes file path and returns the localized path
func (lc *Localizer) localizeFile(path string) (string, error) {
	content, err := lc.ldr.Load(path)
	if err != nil {
		return "", errors.Wrap(err)
	}

	var locPath string
	if loader.IsRemoteFile(path) {
		// TODO(annasong): check if able to add localize directory
		locPath = locFilePath(path)
	} else {
		// ldr has checked that path must be relative; this is subject to change in beta.

		// We must clean path to:
		// 1. avoid symlinks. A `kustomize build` run will fail if we write files to
		//    symlink paths outside the current root, given that we don't want to recreate
		//    the symlinks. Even worse, we could be writing files outside the localize destination.
		// 2. avoid paths that temporarily traverse outside the current root,
		//    i.e. ../../../scope/target/current-root. The localized file will be surrounded by
		//    different directories than its source, and so an uncleaned path may no longer be valid.
		locPath = cleanFilePath(lc.fSys, filesys.ConfirmedDir(lc.ldr.Root()), path)
		// TODO(annasong): check if hits localize directory
	}
	absPath := lc.dst.Join(locPath)
	if err = lc.fSys.MkdirAll(filepath.Dir(absPath)); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create directories to localize file %q", path)
	}
	if err = lc.fSys.WriteFile(absPath, content); err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize file %q", path)
	}
	return locPath, nil
}
