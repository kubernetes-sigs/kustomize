// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/loader"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Args holds localize arguments
type Args struct {
	// target; local copy if remote
	Target filesys.ConfirmedDir

	// directory that bounds target's local references
	// repo directory of local copy if target is remote
	Scope filesys.ConfirmedDir

	// localize destination
	NewDir filesys.ConfirmedDir
}

// Loader is an ifc.Loader that enforces additional constraints specific to kustomize localize.
type Loader struct {
	fSys filesys.FileSystem

	args *Args

	// loader at Loader's current directory
	ifc.Loader

	// whether Loader and all its ancestors are the result of local references
	local bool
}

var _ ifc.Loader = &Loader{}

// NewLoader is the factory method for Loader, under localize constraints, at rawTarget. For invalid localize arguments,
// NewLoader returns an error.
func NewLoader(rawTarget string, rawScope string, rawNewDir string, fSys filesys.FileSystem) (*Loader, Args, error) {
	// check earlier to avoid cleanup
	repoSpec, err := git.NewRepoSpecFromURL(rawTarget)
	if err == nil && repoSpec.Ref == "" {
		return nil, Args{}, errors.Errorf("localize remote root %q missing ref query string parameter", rawTarget)
	}

	// for security, should enforce load restrictions
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, rawTarget, fSys)
	if err != nil {
		return nil, Args{}, errors.WrapPrefixf(err, "unable to establish localize target %q", rawTarget)
	}

	scope, err := establishScope(rawScope, rawTarget, ldr, fSys)
	if err != nil {
		_ = ldr.Cleanup()
		return nil, Args{}, errors.WrapPrefixf(err, "invalid localize scope %q", rawScope)
	}

	newDir, err := createNewDir(rawNewDir, ldr, repoSpec, fSys)
	if err != nil {
		_ = ldr.Cleanup()
		return nil, Args{}, errors.WrapPrefixf(err, "invalid localize destination %q", rawNewDir)
	}

	args := Args{
		Target: filesys.ConfirmedDir(ldr.Root()),
		Scope:  scope,
		NewDir: newDir,
	}
	return &Loader{
		fSys:   fSys,
		args:   &args,
		Loader: ldr,
		local:  scope != "",
	}, args, nil
}

// Load returns the contents of path if path is a valid localize file.
// Otherwise, Load returns an error.
func (ll *Loader) Load(path string) ([]byte, error) {
	// checks in root, and thus in scope
	content, err := ll.Loader.Load(path)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "invalid file reference")
	}
	if !loader.IsRemoteFile(path) && ll.local {
		cleanPath := cleanedRelativePath(ll.fSys, filesys.ConfirmedDir(ll.Root()), path)
		cleanAbs := filepath.Join(ll.Root(), cleanPath)
		dir := filesys.ConfirmedDir(filepath.Dir(cleanAbs))
		// target cannot reference newDir, as this load would've failed prior to localize;
		// not a problem if remote because then reference could only be in newDir if repo copy,
		// which will be cleaned, is inside newDir
		if dir.HasPrefix(ll.args.NewDir) {
			return nil, errors.Errorf("file %q at %q enters localize destination %q", path, cleanAbs, ll.args.NewDir)
		}
	}
	return content, nil
}

// New returns a Loader at path if path is a valid localize root.
// Otherwise, New returns an error.
func (ll *Loader) New(path string) (ifc.Loader, error) {
	ldr, err := ll.Loader.New(path)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "invalid root reference")
	}

	if repo := ldr.Repo(); repo == "" {
		if ll.local && !filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.Scope) {
			return nil, errors.Errorf("root %q outside localize scope %q", ldr.Root(), ll.args.Scope)
		}
		if ll.local && filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.NewDir) {
			return nil, errors.Errorf(
				"root %q references into localize destination %q", ldr.Root(), ll.args.NewDir)
		}
	} else if !hasRef(path) {
		return nil, errors.Errorf("localize remote root %q missing ref query string parameter", path)
	}

	return &Loader{
		fSys:   ll.fSys,
		args:   ll.args,
		Loader: ldr,
		local:  ll.local && ldr.Repo() == "",
	}, nil
}
