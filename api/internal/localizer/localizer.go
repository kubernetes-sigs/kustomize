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
	"sigs.k8s.io/kustomize/kyaml/kio"
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

	kustomization := kt.Kustomization()
	err = lc.localizeNativeFields(&kustomization)
	if err != nil {
		return err
	}
	err = lc.localizeBuiltinPlugins(&kustomization)
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(&kustomization)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to serialize localized kustomization file")
	}
	if err = lc.fSys.WriteFile(lc.dst.Join(konfig.DefaultKustomizationFileName()), content); err != nil {
		return errors.WrapPrefixf(err, "unable to write localized kustomization file")
	}
	return nil
}

// localizeNativeFields localizes paths on kustomize-native fields, like configMapGenerator, that kustomize has a
// built-in understanding of. This excludes helm-related fields, such as `helmGlobals` and `helmCharts`.
func (lc *Localizer) localizeNativeFields(kust *types.Kustomization) error {
	for i := range kust.Patches {
		if kust.Patches[i].Path != "" {
			newPath, err := lc.localizeFile(kust.Patches[i].Path)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize patches path %q", kust.Patches[i].Path)
			}
			kust.Patches[i].Path = newPath
		}
	}
	// TODO(annasong): localize all other kustomization fields: resources, bases, components, crds, configurations,
	// openapi, patchesJson6902, patchesStrategicMerge, replacements, configMapGenerators, secretGenerators
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

// localizeBuiltinPlugins localizes built-in plugins on kust that can contain file paths. The built-in plugins
// can be inline or in a file. This excludes the HelmChartInflationGenerator.
//
// Note that the localization in this function has not been implemented yet.
func (lc *Localizer) localizeBuiltinPlugins(kust *types.Kustomization) error {
	for fieldName, plugins := range map[string]struct {
		entries   []string
		localizer kio.Filter
	}{
		"generators": {
			kust.Generators,
			&localizeBuiltinGenerators{},
		},
		"transformers": {
			kust.Transformers,
			&localizeBuiltinTransformers{},
		},
		"validators": {
			kust.Validators,
			&localizeBuiltinTransformers{},
		},
	} {
		for i, entry := range plugins.entries {
			rm, isPath, err := lc.loadResource(entry)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to load %s entry", fieldName)
			}
			err = rm.ApplyFilter(plugins.localizer)
			if err != nil {
				return errors.Wrap(err)
			}
			localizedPlugin, err := rm.AsYaml()
			if err != nil {
				return errors.WrapPrefixf(err, "unable to serialize localized %s entry %q", fieldName, entry)
			}
			var newEntry string
			if isPath {
				// TODO(annasong): write localizedPlugin to dst
				newEntry = entry
			} else {
				newEntry = string(localizedPlugin)
			}
			plugins.entries[i] = newEntry
		}
	}
	return nil
}

// loadResource tries to load resourceEntry as a file path or inline.
// On success, loadResource returns the loaded resource map and whether resourceEntry is a file path.
func (lc *Localizer) loadResource(resourceEntry string) (resmap.ResMap, bool, error) {
	var fileErr error
	rm, inlineErr := lc.rFactory.NewResMapFromBytes([]byte(resourceEntry))
	if inlineErr != nil {
		rm, fileErr = lc.rFactory.FromFile(lc.ldr, resourceEntry)
		if fileErr != nil {
			err := ResourceLoadError{
				InlineError: inlineErr,
				FileError:   fileErr,
			}
			return nil, false, errors.WrapPrefixf(err, "unable to load resource entry %q", resourceEntry)
		}
	}
	return rm, fileErr == nil, nil
}
