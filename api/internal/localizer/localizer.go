// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	plgnsLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
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
	pLdr      *plgnsLdr.Loader

	// all localize directories created
	localizeDirs map[filesys.ConfirmedDir]struct{}

	// should be LocLoader
	ldr ifc.Loader

	// destination directory in newDir that mirrors ldr's current root.
	dst filesys.ConfirmedDir
}

// NewLocalizer is the factory method for Localizer
func NewLocalizer(ldr *Loader, validator ifc.Validator, rFactory *resmap.Factory, pLdr *plgnsLdr.Loader) (*Localizer, error) {
	toDst, err := filepath.Rel(ldr.args.Scope.String(), ldr.Root())
	if err != nil {
		log.Fatalf("%s: %s", prefixRelErrWhenContains(ldr.args.Scope.String(), ldr.Root()), err.Error())
	}
	dst := ldr.args.NewDir.Join(toDst)
	if err = ldr.fSys.MkdirAll(dst); err != nil {
		return nil, errors.WrapPrefixf(err, "unable to create directory in localize destination")
	}
	return &Localizer{
		fSys:         ldr.fSys,
		validator:    validator,
		rFactory:     rFactory,
		pLdr:         pLdr,
		localizeDirs: make(map[filesys.ConfirmedDir]struct{}),
		ldr:          ldr,
		dst:          filesys.ConfirmedDir(dst),
	}, nil
}

// Localize localizes the root that lt is at
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

// processKust returns a copy of the kustomization at kt with all paths localized.
func (lc *Localizer) processKust(kt *target.KustTarget) (*types.Kustomization, error) {
	kust := kt.Kustomization()

	for name, field := range map[string]*struct {
		Mapper func(string) (string, error)
		Paths  []string
	}{
		"resources": {
			lc.localizePath,
			kust.Resources,
		},
		"components": {
			lc.localizeDir,
			kust.Components,
		},
		"crds": {
			lc.localizeFile,
			kust.Crds,
		},
		"configurations": {
			lc.localizeFile,
			kust.Configurations,
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

	if path, exists := kust.OpenAPI["path"]; exists {
		newPath, err := lc.localizeFile(path)
		if err != nil {
			return &kust, errors.WrapPrefixf(err, "unable to localize openapi path")
		}
		kust.OpenAPI["path"] = newPath
	}

	for i := range kust.ConfigMapGenerator {
		if err := localizeGenerator(lc, &kust.ConfigMapGenerator[i].GeneratorArgs); err != nil {
			return nil, errors.WrapPrefixf(err, "unable to localize configMapGenerator")
		}
	}
	for i := range kust.SecretGenerator {
		if err := localizeGenerator(lc, &kust.SecretGenerator[i].GeneratorArgs); err != nil {
			return nil, errors.WrapPrefixf(err, "unable to localize secretGenerator")
		}
	}
	for name, patches := range map[string][]types.Patch{
		"patches":         kust.Patches,
		"patchesJson6902": kust.PatchesJson6902,
	} {
		for i := range patches {
			if patches[i].Path != "" {
				newPath, err := lc.localizeFile(patches[i].Path)
				if err != nil {
					return nil, errors.WrapPrefixf(err, "unable to localize %s path", name)
				}
				patches[i].Path = newPath
			}
		}
	}
	if err := localizePatchesStrategicMerge(lc, kust.PatchesStrategicMerge); err != nil {
		return nil, err
	}
	if err := localizeReplacements(lc, kust.Replacements); err != nil {
		return nil, err
	}

	return &kust, nil
}

// localizePath localizes path, a root or file, and returns the localized path
func (lc *Localizer) localizePath(path string) (string, error) {
	locPath, err := lc.localizeDir(path)
	if errors.Is(err, InvalidRootError{}) {
		locPath, err = lc.localizeFile(path)
	}
	if err != nil {
		return "", err
	}
	return locPath, nil
}

// localizeFile localizes file path and returns the localized path
func (lc *Localizer) localizeFile(path string) (string, error) {
	content, err := lc.ldr.Load(path)
	if err != nil {
		return "", errors.Wrap(err)
	}

	var locPath string
	if loader.HasRemoteFileScheme(path) {
		if !lc.addLocalizeDir() {
			return "", errors.Errorf("cannot localize remote %q: %w at %q", path, LocalizeDirExistsError{}, lc.ldr.Root())
		}
		lc.localizeDirs[lc.dst] = struct{}{}
		locPath = locFilePath(path)
	} else { // path must be relative; subject to change in beta
		// avoid symlinks; only write file corresponding to actual location in root
		// avoid path that Load() shows to be in root, but may traverse outside
		// temporarily; for example, ../root/config; problematic for rename and
		// relocation
		locPath = cleanFilePath(lc.fSys, filesys.ConfirmedDir(lc.ldr.Root()), path)
		if !lc.guardLocalizeDir(locPath) {
			abs := filepath.Join(lc.ldr.Root(), locPath)
			return "", errors.Errorf("invalid local file path %q at %q: %w", path, abs, LocalizeDirExistsError{})
		}
	}
	cleanPath := lc.dst.Join(locPath)
	if err = lc.fSys.MkdirAll(filepath.Dir(cleanPath)); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create directories to localize file %q", path)
	}
	if err = lc.fSys.WriteFile(cleanPath, content); err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize file %q", path)
	}
	return locPath, nil
}

// localizeDir localizes root path and returns the localized path
func (lc *Localizer) localizeDir(path string) (string, error) {
	ldr, err := lc.ldr.New(path)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer func() { _ = ldr.Cleanup() }()

	var locPath string
	if repo, isRemote := ldr.Repo(); isRemote {
		if !lc.addLocalizeDir() {
			return "", errors.Errorf("cannot localize remote %q: %w at %q", path, LocalizeDirExistsError{}, lc.ldr.Root())
		}
		locPath = locRootPath(path, filesys.ConfirmedDir(repo), filesys.ConfirmedDir(ldr.Root()))
	} else {
		locPath, err = filepath.Rel(lc.ldr.Root(), ldr.Root())
		if err != nil {
			//nolint:gocritic // should never occur, but if hit, something fundamentally wrong; should immediately exit
			log.Fatalf("rel path error for 2 navigable roots %q, %q: %s", lc.ldr.Root(), ldr.Root(), err.Error())
		}
		if !lc.guardLocalizeDir(locPath) {
			return "", errors.Errorf("invalid local root path %q at %q: %w", path, ldr.Root(), LocalizeDirExistsError{})
		}
	}
	newDst := lc.dst.Join(locPath)
	if err = lc.fSys.MkdirAll(newDst); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create root %q in localize destination", path)
	}
	err = (&Localizer{
		fSys:         lc.fSys,
		validator:    lc.validator,
		rFactory:     lc.rFactory,
		pLdr:         lc.pLdr,
		localizeDirs: lc.localizeDirs,
		ldr:          ldr,
		dst:          filesys.ConfirmedDir(newDst),
	}).Localize()
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize root %q", path)
	}
	return locPath, nil
}

// addLocalizeDir returns whether it is able to add a localize directory at dst
func (lc *Localizer) addLocalizeDir() bool {
	if _, exists := lc.localizeDirs[lc.dst]; !exists && lc.fSys.Exists(lc.dst.Join(LocalizeDir)) {
		return false
	}
	lc.localizeDirs[lc.dst] = struct{}{}
	return true
}

// guardLocalizeDir returns false if local path enters a localize directory, and true otherwise
func (lc *Localizer) guardLocalizeDir(path string) bool {
	var prefix string
	for _, dir := range strings.Split(path, string(filepath.Separator)) {
		parent := lc.dst.Join(prefix)
		// if never processed parent, inner directories cannot be localize directories
		if !lc.fSys.Exists(parent) {
			return true
		}
		prefix = filepath.Join(prefix, dir)
		if dir != LocalizeDir {
			continue
		}
		if _, exists := lc.localizeDirs[filesys.ConfirmedDir(parent)]; exists {
			return false
		}
	}
	return true
}
