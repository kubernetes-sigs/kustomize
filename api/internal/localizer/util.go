// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// validateScope returns the scope given localize arguments and targetLdr that corresponds to targetArg
func validateScope(scopeArg string, targetArg string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if _, isRemote := targetLdr.Repo(); isRemote {
		if scopeArg != "" {
			return "", errors.Errorf("scope '%s' specified for remote localize target '%s'", scopeArg, targetArg)
		}
		return "", nil
	}
	// default scope
	if scopeArg == "" {
		return filesys.ConfirmedDir(targetLdr.Root()), nil
	}
	scope, err := filesys.ConfirmDir(fSys, scopeArg)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to establish localize scope")
	}
	if !filesys.ConfirmedDir(targetLdr.Root()).HasPrefix(scope) {
		return scope, errors.Errorf("localize scope '%s' does not contain target '%s' at '%s'",
			scopeArg, targetArg, targetLdr.Root())
	}
	return scope, nil
}

// validateNewDir returns the localize destination directory or error. Note that spec is nil if targetLdr is at local
// target.
func validateNewDir(newDirArg string, targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	var newDirPath string
	if newDirArg == "" {
		newDirPath = defaultNewDir(targetLdr, spec)
	} else {
		newDirPath = newDirArg
	}
	if fSys.Exists(newDirPath) {
		return "", errors.Errorf("localize destination '%s' already exists", newDirPath)
	}
	// destination directory must sit in an existing directory
	if err := fSys.Mkdir(newDirPath); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, newDirPath)
	if err != nil {
		if errCleanup := fSys.RemoveAll(newDir.String()); errCleanup != nil {
			log.Printf("%s", errors.WrapPrefixf(errCleanup, "unable to clean localize destination").Error())
		}
		return "", errors.WrapPrefixf(err, "unable to establish localize destination")
	}

	return newDir, nil
}

// defaultNewDir calculates the default localize destination directory name from targetLdr at the localize target
// and spec of target, which is nil if target is local
func defaultNewDir(targetLdr ifc.Loader, spec *git.RepoSpec) string {
	targetDir := filepath.Base(targetLdr.Root())
	if repo, isRemote := targetLdr.Repo(); isRemote {
		// kustomize doesn't download repo into repo-named folder
		// must find repo folder name from url
		if repo == targetLdr.Root() {
			targetDir = urlBase(spec.OrgRepo)
		}
		return strings.Join([]string{dstPrefix, targetDir, strings.ReplaceAll(spec.Ref, "/", "-")}, "-")
	}
	// special case for local target directory since destination directory cannot have "/" in name
	if targetDir == string(filepath.Separator) {
		return dstPrefix
	}
	return strings.Join([]string{dstPrefix, targetDir}, "-")
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
