// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"bytes"
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
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
// built-in understanding of. This excludes helm.
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
// can be inline or in a file.
func (lc *Localizer) localizeBuiltinPlugins(kust *types.Kustomization) error {
	for fieldName, plugins := range map[string][]string{
		"generator":   kust.Generators,
		"transformer": kust.Transformers,
		"validator":   kust.Validators,
	} {
		for i, entry := range plugins {
			var isPath bool
			rm, inlineErr := lc.rFactory.NewResMapFromBytes([]byte(entry))
			if inlineErr != nil {
				var fileErr error
				rm, fileErr = lc.rFactory.FromFile(lc.ldr, entry)
				if fileErr != nil {
					return errors.WrapPrefixf(fileErr, `unable to localize %s value %q:
when parsing as inline received error: %s
when parsing as filepath received error`, fieldName, entry, inlineErr)
				}
				isPath = true
			}
			localizedPlugin, err := lc.localizePluginEntry(rm)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize %s entry %q", fieldName, entry)
			}
			var newEntry string
			if isPath {
				// TODO(annasong): write localizedPlugin to dst
				newEntry = entry
			} else {
				newEntry = string(localizedPlugin)
			}
			plugins[i] = newEntry
		}
	}
	return nil
}

// localizePluginEntry localizes pluginEntry, the resources in a plugin entry, and returns
// the localized pluginEntry in bytes if they are built-ins that can specify file paths.
// Otherwise, localizePluginEntry returns an error.
//
// Note that the localization in this function has not been implemented yet.
func (lc *Localizer) localizePluginEntry(pluginEntry resmap.ResMap) ([]byte, error) {
	localizedPlugins := make([][]byte, pluginEntry.Size())
	for i, plugin := range pluginEntry.Resources() {
		if !utils.IsBuiltinPlugin(plugin) {
			return nil, errors.Errorf("plugin is not built-in")
		}
		content, err := plugin.AsYAML()
		if err != nil {
			return nil, errors.WrapPrefixf(err, "unable to serialize plugin in entry")
		}
		// TODO(annasong): localize plugins
		localizedPlugins[i] = content
	}
	localizedEntry := bytes.Join(localizedPlugins, []byte("---\n"))
	return bytes.TrimSuffix(localizedEntry, []byte("\n")), nil
}
