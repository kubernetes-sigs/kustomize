// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package loader has a data loading interface and various implementations.
package loader

import (
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// NewLoader returns a Loader pointed at the given target.
// If the target is remote, the loader will be restricted
// to the root and below only.  If the target is local, the
// loader will have the restrictions passed in.  Regardless,
// if a local target attempts to transitively load remote bases,
// the remote bases will all be root-only restricted.
// The loaders below it share their clones and pulls, which
// its Cleanup removes.
func NewLoader(
	lr LoadRestrictorFunc,
	target string, fSys filesys.FileSystem) (ifc.Loader, error) {
	remotes := newRemoteCache()
	ldr, err := newLoaderAtTarget(lr, target, fSys, remotes)
	if err != nil {
		_ = remotes.cleanup()
		return nil, err
	}
	return ldr, nil
}

func newLoaderAtTarget(
	lr LoadRestrictorFunc, target string, fSys filesys.FileSystem,
	remotes *remoteCache) (ifc.Loader, error) {
	repoSpec, err := git.NewRepoSpecFromURL(target)
	if err == nil {
		// The target qualifies as a remote git target.
		return newLoaderAtGitClone(
			repoSpec, fSys, nil, git.ClonerUsingGitExec, oci.PullerUsingRegistry, remotes)
	}
	ociSpec, err := oci.NewRepoSpecFromURL(target)
	if err == nil {
		// The target qualifies as a remote OCI artifact.
		return newLoaderAtOciPull(
			ociSpec, fSys, nil, git.ClonerUsingGitExec, oci.PullerUsingRegistry, remotes)
	}
	if oci.IsOciURL(target) {
		// An oci:// target that does not parse is not a local one either.
		return nil, errors.Wrap(err)
	}
	root, err := filesys.ConfirmDir(fSys, target)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "%s", ErrRtNotDir.Error())
	}
	return newLoaderAtConfirmedDir(
		lr, root, fSys, nil, git.ClonerUsingGitExec, oci.PullerUsingRegistry, remotes), nil
}
