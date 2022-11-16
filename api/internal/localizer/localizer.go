// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Localizer encapsulates all state needed to localize the root at ldr.
type Localizer struct {
	fSys filesys.FileSystem

	// underlying type is Loader
	ldr ifc.Loader

	// destination directory in newDir that mirrors ldr's current root.
	dst filesys.ConfirmedDir

	// kust is the kustomization at ldr's root.
	kust LegacyKust

	// kustFileName stores the name of the kustomization file at ldr root.
	kustFileName string

	rFactory *resmap.Factory
}

// NewLocalizer is the factory method for Localizer
func NewLocalizer(ldr *Loader, rFactory *resmap.Factory) (*Localizer, error) {
	toDst, err := filepath.Rel(ldr.args.Scope.String(), ldr.Root())
	if err != nil {
		log.Fatalf("cannot find path from %q to child directory %q: %s", ldr.args.Scope, ldr.Root(), err)
	}
	dst := ldr.args.NewDir.Join(toDst)
	if err = ldr.fSys.MkdirAll(dst); err != nil {
		return nil, errors.WrapPrefixf(err, "unable to create directory in localize destination")
	}
	return &Localizer{
		fSys:     ldr.fSys,
		ldr:      ldr,
		dst:      filesys.ConfirmedDir(dst),
		rFactory: rFactory,
	}, nil
}

func (lc *Localizer) Localize() error {
	err := lc.loadKust()
	if err != nil {
		return err
	}
	err = lc.processKust()
	if err != nil {
		return err
	}
	locKust, err := lc.kust.Marshal()
	if err != nil {
		return err
	}
	err = lc.fSys.WriteFile(lc.dst.Join(lc.kustFileName), locKust)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to write localized legacy patches")
	}
	return nil
}

// loadKust loads the kustomization that lc is at into lc.legacyKust and lc.legacyPatches
func (lc *Localizer) loadKust() error {
	content, fileName, err := loadKustFile(lc.ldr)
	if err != nil {
		return err
	}
	var legacyKust LegacyKust
	err = (&legacyKust).Unmarshal(content)
	if err != nil {
		return err
	}
	lc.kust = legacyKust
	lc.kustFileName = fileName
	return nil
}

// processKust localizes the paths on lc.kust.
func (lc *Localizer) processKust() error {
	for i := range lc.kust.fields.Patches {
		if lc.kust.fields.Patches[i].Path != "" {
			newPath, err := lc.localizeFile(lc.kust.fields.Patches[i].Path)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize patches path %q", lc.kust.fields.Patches[i].Path)
			}
			lc.kust.fields.Patches[i].Path = newPath
		}
	}
	// TODO(annasong): localize all other kustomization fields: resources, components, crds, configurations,
	// openapi, legacy patches, patchesStrategicMerge, replacements, configMapGenerators, secretGenerators
	// TODO(annasong): localize built-in plugins under generators, transformers, and validators fields
	return nil
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
