// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// newLoaderAtOciURL returns a new Loader for an oci:// build target,
// pulling the artifact from the registry.
func newLoaderAtOciURL(
	target string, fSys filesys.FileSystem) (ifc.Loader, error) {
	repoSpec, err := oci.NewRepoSpecFromURL(target)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return newLoaderAtOciPull(repoSpec, fSys, nil, oci.PullerUsingRegistry)
}

// newAtOciPull returns a new Loader for an oci:// path
// referred to by the current loader.
func (fl *FileLoader) newAtOciPull(path string) (ifc.Loader, error) {
	repoSpec, err := oci.NewRepoSpecFromURL(path)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if err = fl.errIfOciRepoCycle(repoSpec); err != nil {
		return nil, err
	}
	return newLoaderAtOciPull(repoSpec, fl.fSys, fl, fl.puller)
}

// pullerFor returns the puller inherited from the referrer,
// falling back to the registry puller at the top of the chain.
func pullerFor(referrer *FileLoader) oci.Puller {
	if referrer != nil && referrer.puller != nil {
		return referrer.puller
	}
	return oci.PullerUsingRegistry
}

// clonerFor returns the cloner inherited from the referrer,
// falling back to git exec at the top of the chain.
func clonerFor(referrer *FileLoader) git.Cloner {
	if referrer != nil && referrer.cloner != nil {
		return referrer.cloner
	}
	return git.ClonerUsingGitExec
}

// newLoaderAtOciPull returns a new Loader pinned to a temporary
// directory holding a pulled OCI artifact.
func newLoaderAtOciPull(
	repoSpec *oci.RepoSpec, fSys filesys.FileSystem,
	referrer *FileLoader, puller oci.Puller) (ifc.Loader, error) {
	cleaner := repoSpec.Cleaner(fSys)
	err := puller(repoSpec, fSys)
	if err != nil {
		_ = cleaner()
		return nil, err
	}
	root, f, err := fSys.CleanedAbs(repoSpec.AbsPath())
	if err != nil {
		_ = cleaner()
		return nil, errors.Wrap(err)
	}
	// We don't know that the path requested in repoSpec
	// is a directory until we actually pull it and look
	// inside.  That just happened, hence the error check
	// is here.
	if f != "" {
		_ = cleaner()
		return nil, fmt.Errorf(
			"'%s' refers to file '%s'; expecting directory",
			repoSpec.AbsPath(), f)
	}
	// Path in artifact can contain symlinks that exit the artifact.
	// We can only check for this after pulling it.
	if !root.HasPrefix(repoSpec.PullDir()) {
		_ = cleaner()
		return nil, fmt.Errorf("%q refers to directory outside of artifact %q", repoSpec.AbsPath(),
			repoSpec.PullDir())
	}
	return &FileLoader{
		// Pulls never allowed to escape root.
		loadRestrictor: RestrictionRootOnly,
		root:           root,
		referrer:       referrer,
		ociSpec:        repoSpec,
		fSys:           fSys,
		cloner:         clonerFor(referrer),
		puller:         puller,
		cleaner:        cleaner,
	}, nil
}

func (fl *FileLoader) errIfOciContainmentViolation(
	base filesys.ConfirmedDir) error {
	containingRepo := fl.containingOciRepo()
	if containingRepo == nil {
		return nil
	}
	if !base.HasPrefix(containingRepo.PullDir()) {
		return fmt.Errorf(
			"security; bases in kustomizations found in "+
				"pulled oci artifacts must be within the artifact, "+
				"but base '%s' is outside '%s'",
			base, containingRepo.PullDir())
	}
	return nil
}

// Looks back through referrers for an OCI artifact, returning nil
// if none found before reaching a git repo (which is then
// the containing remote root) or the end of the chain.
func (fl *FileLoader) containingOciRepo() *oci.RepoSpec {
	if fl.ociSpec != nil {
		return fl.ociSpec
	}
	if fl.repoSpec != nil || fl.referrer == nil {
		return nil
	}
	return fl.referrer.containingOciRepo()
}

// TODO: Distinguish tags/digests, as for git refs above?
func (fl *FileLoader) errIfOciRepoCycle(newRepoSpec *oci.RepoSpec) error {
	if fl.ociSpec != nil &&
		strings.HasPrefix(fl.ociSpec.Raw(), newRepoSpec.Raw()) {
		return fmt.Errorf(
			"cycle detected: URI '%s' referenced by previous URI '%s'",
			newRepoSpec.Raw(), fl.ociSpec.Raw())
	}
	if fl.referrer == nil {
		return nil
	}
	return fl.referrer.errIfOciRepoCycle(newRepoSpec)
}
