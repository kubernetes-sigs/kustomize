// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const dstPrefix = "localized"

// LocLoader is the Loader for kustomize localize. It Loads under localize constraints.
type LocLoader struct {
	fSys filesys.FileSystem

	// directory that bounds target's local references, empty string if target is remote
	scope filesys.ConfirmedDir

	// localize destination
	newDir filesys.ConfirmedDir

	// loader at LocLoader's current directory
	ifc.Loader

	// whether LocLoader and all its ancestors are the result of local references
	local bool
}

// ValidateLocArgs validates the arguments to kustomize localize.
// On success, ValidateLocArgs returns a LocLoader at targetArg and the localize destination directory; otherwise, an
// error.
func ValidateLocArgs(targetArg string, scopeArg string, newDirArg string, fSys filesys.FileSystem) (*LocLoader, filesys.ConfirmedDir, error) {
	// check earlier to avoid cleanup
	spec, err := git.NewRepoSpecFromURL(targetArg)
	if err == nil && spec.Ref == "" {
		return nil, "", errors.Errorf("localize remote root '%s' missing ref query string parameter", targetArg)
	}

	// for security, should enforce load restrictions
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, targetArg, fSys)
	if err != nil {
		return nil, "", errors.WrapPrefixf(err, "unable to establish localize target '%s'", targetArg)
	}

	scope, err := validateScope(scopeArg, targetArg, ldr, fSys)
	if err != nil {
		if cleanErr := ldr.Cleanup(); cleanErr != nil {
			log.Printf("%s", errors.WrapPrefixf(cleanErr, "unable to clean localize target").Error())
		}
		return nil, "", errors.WrapPrefixf(err, "invalid localize scope '%s'", scopeArg)
	}

	newDir, err := validateNewDir(newDirArg, ldr, spec, fSys)
	if err != nil {
		if cleanErr := ldr.Cleanup(); cleanErr != nil {
			log.Printf("%s", errors.WrapPrefixf(cleanErr, "unable to clean localize target").Error())
		}
		return nil, "", errors.WrapPrefixf(err, "invalid localize destination '%s'", newDirArg)
	}

	return &LocLoader{
		fSys:   fSys,
		scope:  scope,
		newDir: newDir,
		Loader: ldr,
		local:  scope != "",
	}, newDir, nil
}

// Scope returns the localize scope
func (ll *LocLoader) Scope() string {
	return ll.scope.String()
}

// Load returns the contents of path if path is a valid localize file.
// Otherwise, Load returns an error.
func (ll *LocLoader) Load(path string) ([]byte, error) {
	// checks in root, and thus in scope
	content, err := ll.Loader.Load(path)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "invalid file reference")
	}
	if filepath.IsAbs(path) {
		return nil, errors.Errorf("absolute paths not yet supported in alpha: file path '%s' is absolute", path)
	}

	abs := filepath.Join(ll.Root(), path)
	dir, f, err := ll.fSys.CleanedAbs(abs)
	if err != nil {
		// should never happen
		log.Fatalf(errors.WrapPrefixf(err, "cannot clean validated file path '%s'", abs).Error())
	}
	// target cannot reference newDir, as this load would've failed prior to localize;
	// not a problem if remote because then reference could only be in newDir if repo copy,
	// which will be cleaned, is inside newDir
	if ll.local && dir.HasPrefix(ll.newDir) {
		return nil, errors.Errorf("file path '%s' references into localize destination", dir.Join(f))
	}

	return content, nil
}

// New returns a LocLoader at path if path is a valid localize root.
// Otherwise, New returns an error.
func (ll *LocLoader) New(path string) (ifc.Loader, error) {
	spec, err := git.NewRepoSpecFromURL(path)
	if err == nil && spec.Ref == "" {
		return nil, errors.Errorf("localize remote root '%s' missing ref query string parameter", path)
	}

	ldr, err := ll.Loader.New(path)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "invalid root reference")
	}

	var isRemote bool
	if _, isRemote = ldr.Repo(); !isRemote {
		if ll.local && !filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.scope) {
			return nil, errors.Errorf("root '%s' outside localize scope '%s'", ldr.Root(), ll.scope)
		}
		if ll.local && filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.newDir) {
			return nil, errors.Errorf("root '%s' references into localize destination '%s'", ldr.Root(), ll.newDir)
		}
	}

	return &LocLoader{
		fSys:   ll.fSys,
		scope:  ll.scope,
		newDir: ll.newDir,
		Loader: ldr,
		local:  ll.local && !isRemote,
	}, nil
}
