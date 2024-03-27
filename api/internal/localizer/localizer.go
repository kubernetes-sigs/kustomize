// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/generators"
	"sigs.k8s.io/kustomize/api/internal/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
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
// and returns the path to the created newDir.
func Run(target, scope, newDir string, fSys filesys.FileSystem) (string, error) {
	ldr, args, err := NewLoader(target, scope, newDir, fSys)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer func() { _ = ldr.Cleanup() }()

	toDst, err := filepath.Rel(args.Scope.String(), args.Target.String())
	if err != nil {
		log.Panicf("cannot find path from %q to child directory %q: %s", args.Scope, args.Target, err)
	}
	dst := args.NewDir.Join(toDst)
	if err = fSys.MkdirAll(dst); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create directory in localize destination")
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
		return "", errors.WrapPrefixf(err, "unable to localize target %q", target)
	}
	return args.NewDir.String(), nil
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
		return nil, "", errors.Wrap(err)
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
		locPath, err := lc.localizeFile(path)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize openapi path")
		}
		kust.OpenAPI["path"] = locPath
	}

	for fieldName, field := range map[string]struct {
		paths []string
		locFn func(string) (string, error)
	}{
		"bases": {
			// Allow use of deprecated field
			//nolint:staticcheck
			kust.Bases,
			lc.localizeRoot,
		},
		"components": {
			kust.Components,
			lc.localizeRoot,
		},
		"configurations": {
			kust.Configurations,
			lc.localizeFile,
		},
		"crds": {
			kust.Crds,
			lc.localizeFile,
		},
		"resources": {
			kust.Resources,
			lc.localizeResource,
		},
	} {
		for i, path := range field.paths {
			locPath, err := field.locFn(path)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize %s entry", fieldName)
			}
			field.paths[i] = locPath
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
	if err := lc.localizeHelmInflationGenerator(kust); err != nil {
		return err
	}
	if err := lc.localizeHelmCharts(kust); err != nil {
		return err
	}
	if err := lc.localizePatches(kust.Patches); err != nil {
		return errors.WrapPrefixf(err, "unable to localize patches")
	}
	//nolint:staticcheck
	if err := lc.localizePatches(kust.PatchesJson6902); err != nil {
		return errors.WrapPrefixf(err, "unable to localize patchesJson6902")
	}
	//nolint:staticcheck
	for i, patch := range kust.PatchesStrategicMerge {
		locPath, err := lc.localizeK8sResource(string(patch))
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize patchesStrategicMerge entry")
		}
		kust.PatchesStrategicMerge[i] = types.PatchStrategicMerge(locPath)
	}
	for i, replacement := range kust.Replacements {
		locPath, err := lc.localizeFile(replacement.Path)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize replacements entry")
		}
		kust.Replacements[i].Path = locPath
	}
	return nil
}

// localizeGenerator localizes the file paths on generator.
func (lc *localizer) localizeGenerator(generator *types.GeneratorArgs) error {
	locEnvSrc, err := lc.localizeFile(generator.EnvSource)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to localize generator env file")
	}
	locEnvs := make([]string, len(generator.EnvSources))
	for i, env := range generator.EnvSources {
		locEnvs[i], err = lc.localizeFile(env)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator envs file")
		}
	}
	locFiles := make([]string, len(generator.FileSources))
	for i, file := range generator.FileSources {
		locFiles[i], err = lc.localizeFileSource(file)
		if err != nil {
			return err
		}
	}
	generator.EnvSource = locEnvSrc
	generator.EnvSources = locEnvs
	generator.FileSources = locFiles
	return nil
}

// localizeFileSource returns the localized file source found in configMap and
// secretGenerators.
func (lc *localizer) localizeFileSource(source string) (string, error) {
	key, file, err := generators.ParseFileSource(source)
	if err != nil {
		return "", errors.Wrap(err)
	}
	locFile, err := lc.localizeFile(file)
	if err != nil {
		return "", errors.WrapPrefixf(err, "invalid file source %q", source)
	}
	var locSource string
	if source == file {
		locSource = locFile
	} else {
		locSource = key + "=" + locFile
	}
	return locSource, nil
}

// localizeHelmInflationGenerator localizes helmChartInflationGenerator on kust.
// localizeHelmInflationGenerator localizes values files and copies local chart homes.
func (lc *localizer) localizeHelmInflationGenerator(kust *types.Kustomization) error {
	for i, chart := range kust.HelmChartInflationGenerator {
		locFile, err := lc.localizeFile(chart.Values)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize helmChartInflationGenerator entry %d values", i)
		}
		kust.HelmChartInflationGenerator[i].Values = locFile

		locDir, err := lc.copyChartHomeEntry(chart.ChartHome)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to copy helmChartInflationGenerator entry %d", i)
		}
		kust.HelmChartInflationGenerator[i].ChartHome = locDir
	}
	return nil
}

// localizeHelmCharts localizes helmCharts and helmGlobals on kust.
// localizeHelmCharts localizes values files and copies a local chart home.
func (lc *localizer) localizeHelmCharts(kust *types.Kustomization) error {
	for i, chart := range kust.HelmCharts {
		locFile, err := lc.localizeFile(chart.ValuesFile)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize helmCharts entry %d valuesFile", i)
		}
		kust.HelmCharts[i].ValuesFile = locFile

		for j, valuesFile := range chart.AdditionalValuesFiles {
			locFile, err = lc.localizeFile(valuesFile)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize helmCharts entry %d additionalValuesFiles", i)
			}
			kust.HelmCharts[i].AdditionalValuesFiles[j] = locFile
		}
	}
	if kust.HelmGlobals != nil {
		locDir, err := lc.copyChartHomeEntry(kust.HelmGlobals.ChartHome)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to copy helmGlobals")
		}
		kust.HelmGlobals.ChartHome = locDir
	} else if len(kust.HelmCharts) > 0 {
		_, err := lc.copyChartHomeEntry("")
		if err != nil {
			return errors.WrapPrefixf(err, "unable to copy default chart home")
		}
	}
	return nil
}

// localizePatches localizes the file paths on patches if they are non-empty
func (lc *localizer) localizePatches(patches []types.Patch) error {
	for i := range patches {
		locPath, err := lc.localizeFile(patches[i].Path)
		if err != nil {
			return err
		}
		patches[i].Path = locPath
	}
	return nil
}

// localizeResource localizes resource path, a file or root, and returns the
// localized path
func (lc *localizer) localizeResource(path string) (string, error) {
	var locPath string

	content, fileErr := lc.ldr.Load(path)
	// The following check catches the case where path is a repo root.
	// Load on a repo will successfully return its README in HTML.
	// Because HTML does not follow resource formatting, we then correctly try
	// to localize path as a root.
	if fileErr == nil {
		_, resErr := lc.rFactory.NewResMapFromBytes(content)
		if resErr != nil {
			fileErr = errors.WrapPrefixf(resErr, "invalid resource at file %q", path)
		} else {
			locPath, fileErr = lc.localizeFileWithContent(path, content)
		}
	}
	if fileErr != nil {
		var rootErr error
		locPath, rootErr = lc.localizeRoot(path)
		if rootErr != nil {
			err := PathLocalizeError{
				Path:      path,
				FileError: fileErr,
				RootError: rootErr,
			}
			return "", err
		}
	}
	return locPath, nil
}

// localizeFile localizes file path if set and returns the localized path
func (lc *localizer) localizeFile(path string) (string, error) {
	// Some localizable fields can be empty, for example, replacements.path.
	// We rely on the build command to throw errors for the ones that cannot.
	if path == "" {
		return "", nil
	}
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
		if lc.fSys.Exists(lc.root.Join(LocalizeDir)) {
			return "", errors.Errorf("%s already contains %s needed to store file %q", lc.root, LocalizeDir, path)
		}
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
		locPath = cleanedRelativePath(lc.fSys, lc.root, path)
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

// localizeRoot localizes root path if set and returns the localized path
func (lc *localizer) localizeRoot(path string) (string, error) {
	if path == "" {
		return "", nil
	}
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
		if lc.fSys.Exists(lc.root.Join(LocalizeDir)) {
			return "", errors.Errorf("%s already contains %s needed to store root %q", lc.root, LocalizeDir, path)
		}
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

// copyChartHomeEntry copies the helm chart home entry to lc dst
// at the same location relative to the root and returns said relative path.
// If entry is empty, copyChartHomeEntry returns the empty string.
// If entry does not exist, copyChartHome returns entry.
//
// copyChartHomeEntry copies the default home to the same location at dst,
// without following symlinks. An empty entry also indicates the default home.
func (lc *localizer) copyChartHomeEntry(entry string) (string, error) {
	path := entry
	if entry == "" {
		path = types.HelmDefaultHome
	}
	if filepath.IsAbs(path) {
		return "", errors.Errorf("absolute path %q not handled in alpha", path)
	}
	isDefault := lc.root.Join(path) == lc.root.Join(types.HelmDefaultHome)
	locPath, err := lc.copyChartHome(path, !isDefault)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to copy home %q", entry)
	}
	if entry == "" {
		return "", nil
	}
	return locPath, nil
}

// copyChartHome copies path relative to lc root to dst and returns the
// copied location relative to dst. If clean is true, copyChartHome uses path's
// delinked location as the copy destination.
//
// If path does not exist, copyChartHome returns path.
func (lc *localizer) copyChartHome(path string, clean bool) (string, error) {
	path, err := filepath.Rel(lc.root.String(), lc.root.Join(path))
	if err != nil {
		return "", errors.WrapPrefixf(err, "no path to chart home %q", path)
	}
	// Chart home may serve as untar destination.
	// Note that we don't check if path is in scope.
	if !lc.fSys.Exists(lc.root.Join(path)) {
		return path, nil
	}
	// Perform localize directory checks.
	ldr, err := lc.ldr.New(path)
	if err != nil {
		return "", errors.WrapPrefixf(err, "invalid chart home")
	}
	cleaned, err := filesys.ConfirmDir(lc.fSys, ldr.Root())
	if err != nil {
		log.Panicf("unable to confirm validated directory %q: %s", ldr.Root(), err)
	}
	toDst := path
	if clean {
		toDst, err = filepath.Rel(lc.root.String(), cleaned.String())
		if err != nil {
			log.Panicf("no path between scoped directories %q and %q: %s", lc.root, cleaned, err)
		}
	}
	// Note this check does not guarantee that we copied the entire directory.
	if dst := filepath.Join(lc.dst, toDst); !lc.fSys.Exists(dst) {
		err = lc.copyDir(cleaned, filepath.Join(lc.dst, toDst))
		if err != nil {
			return "", errors.WrapPrefixf(err, "unable to copy chart home %q", path)
		}
	}
	return toDst, nil
}

// copyDir copies src to dst. copyDir does not follow symlinks.
func (lc *localizer) copyDir(src filesys.ConfirmedDir, dst string) error {
	err := lc.fSys.Walk(src.String(),
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			pathToCreate, err := filepath.Rel(src.String(), path)
			if err != nil {
				log.Panicf("no path from %q to child file %q: %s", src, path, err)
			}
			pathInDst := filepath.Join(dst, pathToCreate)
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				return nil
			}
			if info.IsDir() {
				err = lc.fSys.MkdirAll(pathInDst)
			} else {
				var content []byte
				content, err = lc.fSys.ReadFile(path)
				if err != nil {
					return errors.Wrap(err)
				}
				err = lc.fSys.WriteFile(pathInDst, content)
			}
			return errors.Wrap(err)
		})
	if err != nil {
		return errors.WrapPrefixf(err, "unable to copy directory %q", src)
	}
	return nil
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

// localizeK8sResource returns the localized resourceEntry if it is a file
// containing a kubernetes resource.
// localizeK8sResource returns resourceEntry if it is an inline resource.
func (lc *localizer) localizeK8sResource(resourceEntry string) (string, error) {
	_, isFile, err := lc.loadK8sResource(resourceEntry)
	if err != nil {
		return "", err
	}
	if isFile {
		return lc.localizeFile(resourceEntry)
	}
	return resourceEntry, nil
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
