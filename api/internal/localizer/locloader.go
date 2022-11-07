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

// LocArgs holds localize arguments
type LocArgs struct {
	// target; local copy if remote
	Target filesys.ConfirmedDir

	// directory that bounds target's local references, empty string if target is remote
	Scope filesys.ConfirmedDir

	// localize destination
	NewDir filesys.ConfirmedDir
}

// locLoader is the Loader for kustomize localize. It is an ifc.Loader that enforces localize constraints.
type locLoader struct {
	fSys filesys.FileSystem

	args *LocArgs

	// loader at locLoader's current directory
	ifc.Loader

	// whether locLoader and all its ancestors are the result of local references
	local bool
}

var _ ifc.Loader = &locLoader{}

// NewLocLoader is the factory method for Loader, under localize constraints, at targetArg. For invalid localize arguments,
// NewLocLoader returns an error.
func NewLocLoader(targetArg string, scopeArg string, newDirArg string, fSys filesys.FileSystem) (ifc.Loader, LocArgs, error) {
	// check earlier to avoid cleanup
	repoSpec, err := git.NewRepoSpecFromURL(targetArg)
	if err == nil && repoSpec.Ref == "" {
		return nil, LocArgs{},
			errors.Errorf("localize remote root '%s' missing ref query string parameter", targetArg)
	}

	// for security, should enforce load restrictions
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, targetArg, fSys)
	if err != nil {
		return nil, LocArgs{}, errors.WrapPrefixf(err, "unable to establish localize target '%s'", targetArg)
	}

	scope, err := establishScope(scopeArg, targetArg, ldr, fSys)
	if err != nil {
		_ = ldr.Cleanup()
		return nil, LocArgs{}, errors.WrapPrefixf(err, "invalid localize scope '%s'", scopeArg)
	}

	newDir, err := createNewDir(newDirArg, ldr, repoSpec, fSys)
	if err != nil {
		_ = ldr.Cleanup()
		return nil, LocArgs{}, errors.WrapPrefixf(err, "invalid localize destination '%s'", newDirArg)
	}

	args := LocArgs{
		Target: filesys.ConfirmedDir(ldr.Root()),
		Scope:  scope,
		NewDir: newDir,
	}
	return &locLoader{
		fSys:   fSys,
		args:   &args,
		Loader: ldr,
		local:  scope != "",
	}, args, nil
}

// Load returns the contents of path if path is a valid localize file.
// Otherwise, Load returns an error.
func (ll *locLoader) Load(path string) ([]byte, error) {
	// checks in root, and thus in scope
	content, err := ll.Loader.Load(path)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "invalid file reference")
	}
	if filepath.IsAbs(path) {
		return nil, errors.Errorf("absolute paths not yet supported in alpha: file path '%s' is absolute", path)
	}
	if ll.local {
		abs := filepath.Join(ll.Root(), path)
		dir, f, err := ll.fSys.CleanedAbs(abs)
		if err != nil {
			// should never happen
			log.Fatalf(errors.WrapPrefixf(err, "cannot clean validated file path '%s'", abs).Error())
		}
		// target cannot reference newDir, as this load would've failed prior to localize;
		// not a problem if remote because then reference could only be in newDir if repo copy,
		// which will be cleaned, is inside newDir
		if dir.HasPrefix(ll.args.NewDir) {
			return nil, errors.Errorf(
				"file path '%s' references into localize destination '%s'", dir.Join(f), ll.args.NewDir)
		}
	}
	return content, nil
}

// New returns a Loader at path if path is a valid localize root.
// Otherwise, New returns an error.
func (ll *locLoader) New(path string) (ifc.Loader, error) {
	ldr, err := ll.Loader.New(path)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "invalid root reference")
	}

	if repo := ldr.Repo(); repo == "" {
		if ll.local && !filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.Scope) {
			return nil, errors.Errorf("root '%s' outside localize scope '%s'", ldr.Root(), ll.args.Scope)
		}
		if ll.local && filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.NewDir) {
			return nil, errors.Errorf(
				"root '%s' references into localize destination '%s'", ldr.Root(), ll.args.NewDir)
		}
	} else if !hasRef(path) {
		return nil, errors.Errorf("localize remote root '%s' missing ref query string parameter", path)
	}

	return &locLoader{
		fSys:   ll.fSys,
		args:   ll.args,
		Loader: ldr,
		local:  ll.local && ldr.Repo() == "",
	}, nil
}
