// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// DstPrefix prefixes the target and ref, if target is remote, in the default localize destination directory name
const DstPrefix = "localized"

// LocalizeDir is the name of the localize directories used to store remote content in the localize destination
const LocalizeDir = "localized-files"

// establishScope returns the effective scope given localize arguments and targetLdr at rawTarget. For remote rawTarget,
// the effective scope is the downloaded repo.
func establishScope(rawScope string, rawTarget string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if repo := targetLdr.Repo(); repo != "" {
		if rawScope != "" {
			return "", errors.Errorf("scope %q specified for remote localize target %q", rawScope, rawTarget)
		}
		return filesys.ConfirmedDir(repo), nil
	}
	// default scope
	if rawScope == "" {
		return filesys.ConfirmedDir(targetLdr.Root()), nil
	}
	scope, err := filesys.ConfirmDir(fSys, rawScope)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to establish localize scope")
	}
	if !filesys.ConfirmedDir(targetLdr.Root()).HasPrefix(scope) {
		return scope, errors.Errorf("localize scope %q does not contain target %q at %q", rawScope, rawTarget,
			targetLdr.Root())
	}
	return scope, nil
}

// createNewDir returns the localize destination directory or error. Note that spec is nil if targetLdr is at local
// target.
func createNewDir(rawNewDir string, targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if rawNewDir == "" {
		rawNewDir = defaultNewDir(targetLdr, spec)
	}
	if fSys.Exists(rawNewDir) {
		return "", errors.Errorf("localize destination %q already exists", rawNewDir)
	}
	// destination directory must sit in an existing directory
	if err := fSys.Mkdir(rawNewDir); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, rawNewDir)
	if err != nil {
		if errCleanup := fSys.RemoveAll(newDir.String()); errCleanup != nil {
			log.Printf("%s", errors.WrapPrefixf(errCleanup, "unable to clean localize destination"))
		}
		return "", errors.WrapPrefixf(err, "unable to establish localize destination")
	}

	return newDir, nil
}

// defaultNewDir calculates the default localize destination directory name from targetLdr at the localize target
// and spec of target, which is nil if target is local
func defaultNewDir(targetLdr ifc.Loader, spec *git.RepoSpec) string {
	targetDir := filepath.Base(targetLdr.Root())
	if repo := targetLdr.Repo(); repo != "" {
		// kustomize doesn't download repo into repo-named folder
		// must find repo folder name from url
		if repo == targetLdr.Root() {
			targetDir = urlBase(spec.OrgRepo)
		}
		return strings.Join([]string{DstPrefix, targetDir, strings.ReplaceAll(spec.Ref, "/", "-")}, "-")
	}
	// special case for local target directory since destination directory cannot have "/" in name
	if targetDir == string(filepath.Separator) {
		return DstPrefix
	}
	return strings.Join([]string{DstPrefix, targetDir}, "-")
}

// urlBase is the url equivalent of filepath.Base
func urlBase(url string) string {
	cleaned := strings.TrimRight(url, "/")
	i := strings.LastIndex(cleaned, "/")
	if i < 0 {
		return cleaned
	}
	return cleaned[i+1:]
}

// hasRef checks if repoURL has ref query string parameter
func hasRef(repoURL string) bool {
	repoSpec, err := git.NewRepoSpecFromURL(repoURL)
	if err != nil {
		log.Fatalf("unable to parse validated root url: %s", err)
	}
	return repoSpec.Ref != ""
}

// cleanFilePath returns file cleaned, where file is a relative path to root on fSys
func cleanFilePath(fSys filesys.FileSystem, root filesys.ConfirmedDir, file string) string {
	abs := root.Join(file)
	dir, f, err := fSys.CleanedAbs(abs)
	if err != nil {
		log.Fatalf("cannot clean validated file path %q: %s", abs, err)
	}
	locPath, err := filepath.Rel(root.String(), dir.Join(f))
	if err != nil {
		log.Fatalf("cannot find path from parent %q to file %q: %s", root, dir.Join(f), err)
	}
	return locPath
}

// locFilePath returns the relative localized path of validated file url fileURL
// TODO(annasong): implement
func locFilePath(_ string) string {
	return filepath.Join(LocalizeDir, "")
}

func isBuiltinPlugin(res *resource.Resource) bool {
	return res.GetGvk().Group == "" && res.GetGvk().Version == konfig.BuiltinPluginApiVersion
}
