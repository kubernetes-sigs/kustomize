// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/generators"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// localizer encapsulates all state needed to localize the root at ldr.
type localizer struct {
	fSys filesys.FileSystem

	// underlying type is Loader
	ldr ifc.Loader

	// root is at ldr.Root()
	root filesys.ConfirmedDir

	rFactory *resmap.Factory

	// destination directory in newDir that mirrors root
	dst string
}

// Run attempts to localize the kustomization root at target with the given localize arguments
func Run(target string, scope string, newDir string, fSys filesys.FileSystem) error {
	ldr, args, err := NewLoader(target, scope, newDir, fSys)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() { _ = ldr.Cleanup() }()

	toDst, err := filepath.Rel(args.Scope.String(), args.Target.String())
	if err != nil {
		log.Panicf("cannot find path from %q to child directory %q: %s", args.Scope, args.Target, err)
	}
	dst := args.NewDir.Join(toDst)
	if err = fSys.MkdirAll(dst); err != nil {
		return errors.WrapPrefixf(err, "unable to create directory in localize destination")
	}

	err = (&localizer{
		fSys:     fSys,
		ldr:      ldr,
		root:     args.Target,
		rFactory: resmap.NewFactory(provider.NewDepProvider().GetResourceFactory()),
		dst:      dst,
	}).localize()
	if err != nil {
		errCleanup := fSys.RemoveAll(args.NewDir.String())
		if errCleanup != nil {
			log.Printf("unable to clean localize destination: %s", errCleanup)
		}
		return errors.WrapPrefixf(err, "unable to localize target %q", target)
	}
	return nil
}

// localize localizes the root that lc is at
func (lc *localizer) localize() error {
	kustomization, kustFileName, err := lc.load()
	if err != nil {
		return err
	}
	err = lc.localizeNativeFields(kustomization)
	if err != nil {
		return err
	}
	err = lc.localizeBuiltinPlugins(kustomization)
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(kustomization)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to serialize localized kustomization file")
	}
	if err = lc.fSys.WriteFile(filepath.Join(lc.dst, kustFileName), content); err != nil {
		return errors.WrapPrefixf(err, "unable to write localized kustomization file")
	}
	return nil
}

// load returns the kustomization at lc.root and the file name under which it was found
func (lc *localizer) load() (*types.Kustomization, string, error) {
	content, kustFileName, err := target.LoadKustFile(lc.ldr)
	if err != nil {
		return nil, "", errors.Wrap(err)
	}

	var kust types.Kustomization
	err = (&kust).Unmarshal(content)
	if err != nil {
		return nil, "", errors.WrapPrefixf(err, "invalid kustomization")
	}

	// Localize intentionally does not replace legacy fields to return a localized kustomization
	// with as much resemblance to the original as possible.
	// Localize also intentionally does not enforce fields, as localize does not wish to unnecessarily
	// repeat the responsibilities of kustomize build.

	return &kust, kustFileName, nil
}

// localizeNativeFields localizes paths on kustomize-native fields, like configMapGenerator, that kustomize has a
// built-in understanding of. This excludes helm-related fields, such as `helmGlobals` and `helmCharts`.
func (lc *localizer) localizeNativeFields(kust *types.Kustomization) error {
	if path, exists := kust.OpenAPI["path"]; exists {
		newPath, err := lc.localizeFile(path)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize openapi path")
		}
		kust.OpenAPI["path"] = newPath
	}

	for fieldName, field := range map[string]struct {
		paths []string
		locFn func(string) (string, error)
	}{
		"components": {
			kust.Components,
			lc.localizeDir,
		},
		"configurations": {
			kust.Configurations,
			lc.localizeFile,
		},
		"crds": {
			kust.Crds,
			lc.localizeFile,
		},
	} {
		for i, path := range field.paths {
			newPath, err := field.locFn(path)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize %s path", fieldName)
			}
			field.paths[i] = newPath
		}
	}

	for i := range kust.ConfigMapGenerator {
		if err := lc.localizeGenerator(&kust.ConfigMapGenerator[i].GeneratorArgs); err != nil {
			return errors.WrapPrefixf(err, "unable to localize configMapGenerator")
		}
	}
	for i := range kust.SecretGenerator {
		if err := lc.localizeGenerator(&kust.SecretGenerator[i].GeneratorArgs); err != nil {
			return errors.WrapPrefixf(err, "unable to localize secretGenerator")
		}
	}
	if err := lc.localizePatches(kust.Patches); err != nil {
		return errors.WrapPrefixf(err, "unable to localize patches")
	}
	// Allow use of deprecated field
	//nolint:staticcheck
	if err := lc.localizePatches(kust.PatchesJson6902); err != nil {
		return errors.WrapPrefixf(err, "unable to localize patchesJson6902")
	}
	//nolint:staticcheck
	for i, patch := range kust.PatchesStrategicMerge {
		localizedPath, err := lc.localizeK8sResource(string(patch))
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize patchesStrategicMerge entry")
		}
		if localizedPath != "" {
			kust.PatchesStrategicMerge[i] = types.PatchStrategicMerge(localizedPath)
		}
	}
	for i, replacement := range kust.Replacements {
		if replacement.Path != "" {
			newPath, err := lc.localizeFile(replacement.Path)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize replacements entry")
			}
			kust.Replacements[i].Path = newPath
		}
	}

	// TODO(annasong): localize all other kustomization fields: resources, bases, configMapGenerator.env, secretGenerator.env
	return nil
}

// localizeGenerator localizes the file paths on generator.
func (lc *localizer) localizeGenerator(generator *types.GeneratorArgs) error {
	locEnvs := make([]string, len(generator.EnvSources))
	for i, env := range generator.EnvSources {
		newPath, err := lc.localizeFile(env)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator envs file")
		}
		locEnvs[i] = newPath
	}
	locFiles := make([]string, len(generator.FileSources))
	for i, file := range generator.FileSources {
		k, f, err := generators.ParseFileSource(file)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to parse generator files entry %q", file)
		}
		newFile, err := lc.localizeFile(f)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator files path in entry %q", file)
		}
		if f != file {
			newFile = k + "=" + newFile
		}
		locFiles[i] = newFile
	}
	generator.EnvSources = locEnvs
	generator.FileSources = locFiles
	return nil
}

// localizePatches localizes the file paths on patches if they are non-empty
func (lc *localizer) localizePatches(patches []types.Patch) error {
	for i := range patches {
		if patches[i].Path != "" {
			newPath, err := lc.localizeFile(patches[i].Path)
			if err != nil {
				return err
			}
			patches[i].Path = newPath
		}
	}
	return nil
}

// localizeFile localizes file path and returns the localized path
func (lc *localizer) localizeFile(path string) (string, error) {
	content, err := lc.ldr.Load(path)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return lc.localizeFileWithContent(path, content)
}

// localizeFileWithContent writes content to the localized file path and returns the localized path.
func (lc *localizer) localizeFileWithContent(path string, content []byte) (string, error) {
	var locPath string
	if loader.IsRemoteFile(path) {
		// TODO(annasong): You need to check if you can add a localize directory here to store
		// the remote file content. There may be a directory that shares the localize directory name.
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
		locPath = cleanFilePath(lc.fSys, lc.root, path)
	}
	absPath := filepath.Join(lc.dst, locPath)
	if err := lc.fSys.MkdirAll(filepath.Dir(absPath)); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create directories to localize file %q", path)
	}
	if err := lc.fSys.WriteFile(absPath, content); err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize file %q", path)
	}
	return locPath, nil
}

// localizeDir localizes root path and returns the localized path
func (lc *localizer) localizeDir(path string) (string, error) {
	ldr, err := lc.ldr.New(path)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer func() { _ = ldr.Cleanup() }()

	root, err := filesys.ConfirmDir(lc.fSys, ldr.Root())
	if err != nil {
		log.Panicf("unable to establish validated root reference %q: %s", path, err)
	}
	var locPath string
	if repo := ldr.Repo(); repo != "" {
		// TODO(annasong): You need to check if you can add a localize directory here to store
		// the remote file content. There may be a directory that shares the localize directory name.
		locPath, err = locRootPath(path, repo, root, lc.fSys)
		if err != nil {
			return "", err
		}
	} else {
		locPath, err = filepath.Rel(lc.root.String(), root.String())
		if err != nil {
			log.Panicf("cannot find relative path between scoped localize roots %q and %q: %s", lc.root, root, err)
		}
	}
	newDst := filepath.Join(lc.dst, locPath)
	if err = lc.fSys.MkdirAll(newDst); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create root %q in localize destination", path)
	}
	err = (&localizer{
		fSys:     lc.fSys,
		ldr:      ldr,
		root:     root,
		rFactory: lc.rFactory,
		dst:      newDst,
	}).localize()
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to localize root %q", path)
	}
	return locPath, nil
}

// localizeBuiltinPlugins localizes built-in plugins on kust that can contain file paths. The built-in plugins
// can be inline or in a file. This excludes the HelmChartInflationGenerator.
//
// Note that the localization in this function has not been implemented yet.
func (lc *localizer) localizeBuiltinPlugins(kust *types.Kustomization) error {
	for fieldName, entries := range map[string][]string{
		"generators":   kust.Generators,
		"transformers": kust.Transformers,
		"validators":   kust.Validators,
	} {
		for i, entry := range entries {
			rm, isPath, err := lc.loadK8sResource(entry)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to load %s entry", fieldName)
			}
			err = rm.ApplyFilter(&localizeBuiltinPlugins{lc: lc})
			if err != nil {
				return errors.Wrap(err)
			}
			localizedPlugin, err := rm.AsYaml()
			if err != nil {
				return errors.WrapPrefixf(err, "unable to serialize localized %s entry %q", fieldName, entry)
			}
			var localizedEntry string
			if isPath {
				localizedEntry, err = lc.localizeFileWithContent(entry, localizedPlugin)
				if err != nil {
					return errors.WrapPrefixf(err, "unable to localize %s entry", fieldName)
				}
			} else {
				localizedEntry = string(localizedPlugin)
			}
			entries[i] = localizedEntry
		}
	}
	return nil
}

// localizeK8sResource returns the localized file path if resourceEntry is a
// file containing a kubernetes resource.
// localizeK8sResource returns the empty string if resourceEntry is an inline
// resource.
func (lc *localizer) localizeK8sResource(resourceEntry string) (string, error) {
	_, isFile, err := lc.loadK8sResource(resourceEntry)
	if err != nil {
		return "", err
	}
	if isFile {
		return lc.localizeFile(resourceEntry)
	}
	return "", nil
}

// loadK8sResource tries to load resourceEntry as a file path or inline
// kubernetes resource.
// On success, loadK8sResource returns the loaded resource map and whether
// resourceEntry is a file path.
func (lc *localizer) loadK8sResource(resourceEntry string) (resmap.ResMap, bool, error) {
	rm, inlineErr := lc.rFactory.NewResMapFromBytes([]byte(resourceEntry))
	if inlineErr != nil {
		var fileErr error
		rm, fileErr = lc.rFactory.FromFile(lc.ldr, resourceEntry)
		if fileErr != nil {
			err := ResourceLoadError{
				InlineError: inlineErr,
				FileError:   fileErr,
			}
			return nil, false, errors.WrapPrefixf(err, "unable to load resource entry %q", resourceEntry)
		}
	}
	return rm, inlineErr != nil, nil
}
