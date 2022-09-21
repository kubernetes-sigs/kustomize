// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	pluginsLoader "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// localizer localizes the kustomization root at ldr.
type localizer struct {
	fSys filesys.FileSystem

	ldr ifc.Loader

	// destination directory in newDir that mirrors ldr's current root.
	dst filesys.ConfirmedDir

	// all localize directories created
	localizeDirs map[filesys.ConfirmedDir]struct{}

	// kusttarget fields
	rFactory  *resmap.Factory
	pLdr      *pluginsLoader.Loader
	validator ifc.Validator
}

// Run attempts to localize the kustomization root at targetArg with the given localize arguments
func Run(fSys filesys.FileSystem, targetArg string, scopeArg string, newDirArg string) error {
	ldr, args, err := NewLocLoader(targetArg, scopeArg, newDirArg, fSys)
	if err != nil {
		return err
	}
	defer func() { _ = ldr.Cleanup() }()

	toDst, err := filepath.Rel(args.Scope.String(), ldr.Root())
	if err != nil {
		//nolint:gocritic // should never happen, but if hit, something fundamentally wrong; should immediately exit
		log.Fatalf("%s: %s", prefixRelErrWhenContains(args.Scope.String(), ldr.Root()), err.Error())
	}
	dst := args.NewDir.Join(toDst)
	if err = fSys.MkdirAll(dst); err != nil {
		return errors.WrapPrefixf(err, "unable to create directory in localize destination")
	}

	depProvider := provider.NewDepProvider()
	rFactory := resmap.NewFactory(depProvider.GetResourceFactory())
	err = (&localizer{
		fSys:         fSys,
		ldr:          ldr,
		dst:          filesys.ConfirmedDir(dst),
		localizeDirs: map[filesys.ConfirmedDir]struct{}{},
		validator:    depProvider.GetFieldValidator(),
		rFactory:     rFactory,
		pLdr:         pluginsLoader.NewLoader(krusty.MakeDefaultOptions().PluginConfig, rFactory, filesys.MakeFsOnDisk()),
	}).localize()
	if err != nil {
		if errCleanup := fSys.RemoveAll(args.NewDir.String()); errCleanup != nil {
			log.Printf("unable to clean localize destination: %s", errCleanup.Error())
		}
		return errors.WrapPrefixf(err, "unable to localize target '%s'", targetArg)
	}
	return nil
}

// localize localizes the root that lclzr is at
func (lclzr *localizer) localize() error {
	kt := target.NewKustTarget(lclzr.ldr, lclzr.validator, lclzr.rFactory, lclzr.pLdr)
	err := kt.Load()
	if err != nil {
		return errors.Wrap(err)
	}

	kust, err := lclzr.processKust(kt)
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(kust)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to serialize localized kustomization file")
	}
	if err = lclzr.fSys.WriteFile(lclzr.dst.Join(konfig.DefaultKustomizationFileName()), content); err != nil {
		return errors.WrapPrefixf(err, "unable to write localized kustomization file")
	}
	return nil
}

// processKust returns a copy of the kustomization at kt with all paths localized.
func (lclzr *localizer) processKust(kt *target.KustTarget) (*types.Kustomization, error) {
	kust := kt.Kustomization()

	if path, exists := kust.OpenAPI["path"]; exists {
		var err error
		kust.OpenAPI["path"], err = lclzr.localizeFile(path)
		if err != nil {
			return &kust, errors.WrapPrefixf(err, "unable to localize openapi path")
		}
	}

	for i, p := range kust.PatchesStrategicMerge {
		if _, err := lclzr.rFactory.RF().SliceFromBytes([]byte(p)); err == nil { // inline
			continue
		}
		newPath, err := lclzr.localizeFile(string(p)) // file
		if err != nil {
			return &kust, errors.WrapPrefixf(err, "unable to localize patchesStrategicMerge path")
		}
		kust.PatchesStrategicMerge[i] = types.PatchStrategicMerge(newPath)
	}
	for name, patches := range map[string][]types.Patch{
		"patchesJson6902": kust.PatchesJson6902,
		"patches":         kust.Patches,
	} {
		for i, p := range patches {
			if p.Path == "" {
				continue
			}
			newPath, err := lclzr.localizeFile(p.Path)
			if err != nil {
				return &kust, errors.WrapPrefixf(err, "unable to localize %s path", name)
			}
			patches[i].Path = newPath
		}
	}

	for i, r := range kust.Replacements {
		if r.Path == "" {
			continue
		}
		newPath, err := lclzr.localizeFile(r.Path)
		if err != nil {
			return &kust, errors.WrapPrefixf(err, "unable to localize replacements path")
		}
		kust.Replacements[i].Path = newPath
	}

	for name, field := range map[string]*struct {
		Mapper func(string) (string, error)
		Paths  []string
	}{
		"resources": {
			lclzr.localizePath,
			kust.Resources,
		},
		"components": {
			lclzr.localizeDir,
			kust.Components,
		},
		"crds": {
			lclzr.localizeFile,
			kust.Crds,
		},
	} {
		for i, path := range field.Paths {
			newPath, err := field.Mapper(path)
			if err != nil {
				return &kust, errors.WrapPrefixf(err, "unable to localize %s path", name)
			}
			field.Paths[i] = newPath
		}
	}

	for i := 0; i < len(kust.ConfigMapGenerator); i++ {
		if err := localizeGenerator(&kust.ConfigMapGenerator[i].GeneratorArgs, lclzr); err != nil {
			return &kust, errors.WrapPrefixf(err, "unable to localize configMapGenerator path")
		}
	}
	for i := 0; i < len(kust.SecretGenerator); i++ {
		if err := localizeGenerator(&kust.SecretGenerator[i].GeneratorArgs, lclzr); err != nil {
			return &kust, errors.WrapPrefixf(err, "unable to localize secretGenerator path")
		}
	}

	if len(kust.Generators) > 0 {
		log.Printf("'%v' %s", kust.Generators, GeneratorWarning)
	}
	if len(kust.Transformers) > 0 {
		log.Printf("'%v' %s", kust.Transformers, TransformerWarning)
	}

	return &kust, nil
}

// localizePath localizes path, a root or file, and returns the localized path
func (lclzr *localizer) localizePath(path string) (string, error) {
	locPath, err := lclzr.localizeDir(path)
	if errors.Is(err, ErrInvalidRoot) {
		locPath, err = lclzr.localizeFile(path)
	}
	if err != nil {
		return "", err
	}
	return locPath, nil
}

// localizeFile localizes file path and returns the localized path
func (lclzr *localizer) localizeFile(path string) (string, error) {
	content, err := lclzr.ldr.Load(path)
	if err != nil {
		return "", errors.Wrap(err)
	}

	var locPath string
	if loader.HasRemoteFileScheme(path) {
		if !lclzr.addLocalizeDir() {
			return "", errors.Errorf("cannot localize remote '%s': %w at '%s'", path, ErrLocalizeDirExists, lclzr.ldr.Root())
		}
		lclzr.localizeDirs[lclzr.dst] = struct{}{}
		locPath = locFilePath(path)
	} else { // path must be relative; subject to change in beta
		// avoid symlinks; only write file corresponding to actual location in root
		// avoid path that Load() shows to be in root, but may traverse outside
		// temporarily; for example, ../root/config; problematic for rename and
		// relocation
		locPath = cleanFilePath(lclzr.fSys, filesys.ConfirmedDir(lclzr.ldr.Root()), path)
		if !lclzr.guardLocalizeDir(locPath) {
			abs := filepath.Join(lclzr.ldr.Root(), locPath)
			return "", errors.Errorf("invalid local file path '%s' at '%s': %w", path, abs, ErrLocalizeDirExists)
		}
	}
	cleanPath := lclzr.dst.Join(locPath)
	if err = lclzr.fSys.MkdirAll(filepath.Dir(cleanPath)); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create directories to localize file '%s'", path)
	}
	if err = lclzr.fSys.WriteFile(cleanPath, content); err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize file '%s'", path)
	}
	return locPath, nil
}

// localizeDir localizes root path and returns the localized path
func (lclzr *localizer) localizeDir(path string) (string, error) {
	ldr, err := lclzr.ldr.New(path)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer func() { _ = ldr.Cleanup() }()

	var locPath string
	if repo, isRemote := ldr.Repo(); isRemote {
		if !lclzr.addLocalizeDir() {
			return "", errors.Errorf("cannot localize remote '%s': %w at '%s'", path, ErrLocalizeDirExists, lclzr.ldr.Root())
		}
		locPath = locRootPath(path, filesys.ConfirmedDir(repo), filesys.ConfirmedDir(ldr.Root()))
	} else {
		locPath, err = filepath.Rel(lclzr.ldr.Root(), ldr.Root())
		if err != nil {
			//nolint:gocritic // should never occur, but if hit, something fundamentally wrong; should immediately exit
			log.Fatalf("rel path error for 2 navigable roots '%s', '%s': %s", lclzr.ldr.Root(), ldr.Root(), err.Error())
		}
		if !lclzr.guardLocalizeDir(locPath) {
			return "", errors.Errorf("invalid local root path '%s' at '%s': %w", path, ldr.Root(), ErrLocalizeDirExists)
		}
	}
	newDst := lclzr.dst.Join(locPath)
	if err = lclzr.fSys.MkdirAll(newDst); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create root '%s' in localize destination", path)
	}
	err = (&localizer{
		fSys:         lclzr.fSys,
		ldr:          ldr,
		dst:          filesys.ConfirmedDir(newDst),
		localizeDirs: lclzr.localizeDirs,
		validator:    lclzr.validator,
		rFactory:     lclzr.rFactory,
		pLdr:         lclzr.pLdr,
	}).localize()
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize root '%s'", path)
	}
	return locPath, nil
}

// addLocalizeDir returns whether it is able to add a localize directory at dst
func (lclzr *localizer) addLocalizeDir() bool {
	if _, exists := lclzr.localizeDirs[lclzr.dst]; !exists && lclzr.fSys.Exists(lclzr.dst.Join(LocalizeDir)) {
		return false
	}
	lclzr.localizeDirs[lclzr.dst] = struct{}{}
	return true
}

// guardLocalizeDir returns false if local path enters a localize directory, and true otherwise
func (lclzr *localizer) guardLocalizeDir(path string) bool {
	var prefix string
	for _, dir := range strings.Split(path, string(filepath.Separator)) {
		parent := lclzr.dst.Join(prefix)
		// if never processed parent, inner directories cannot be localize directories
		if !lclzr.fSys.Exists(parent) {
			return true
		}
		prefix = filepath.Join(prefix, dir)
		if dir != LocalizeDir {
			continue
		}
		if _, exists := lclzr.localizeDirs[filesys.ConfirmedDir(parent)]; exists {
			return false
		}
	}
	return true
}
