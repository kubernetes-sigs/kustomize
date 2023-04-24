// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package loader has a data loading interface and various implementations.
package loader

import (
	"fmt"

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
func NewLoader(
	lr LoadRestrictorFunc,
	target string, fSys filesys.FileSystem) (ifc.Loader, error) {
	// target is a Git URL
	repoSpec, err := git.NewRepoSpecFromURL(target)
	if err == nil {
		// The target qualifies as a remote git target.
		return newLoaderAtGitClone(
			repoSpec, fSys, nil, git.ClonerUsingGitExec)
	}
	// target is an OCI endpoint
	// TODO: Implement OCI support
	ociSpec, err := oci.NewOCISpecFromURL(target)
	fmt.Printf("Spec: %#v", ociSpec)
	if err == nil {
		return nil, errors.Errorf("not yet implemented")
		// return newLoaderAtOCIManifest(
		// 	ociSpec, fSys, nil, oci.DownloadImg)
	}

	// target is a local filesystem
	root, err := filesys.ConfirmDir(fSys, target)
	if err != nil {
		return nil, errors.WrapPrefixf(err, ErrRtNotDir.Error())
	}
	return newLoaderAtConfirmedDir(
		lr, root, fSys, nil, git.ClonerUsingGitExec), nil
}
