// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package loader has a data loading interface and various implementations.
package loader

import (
	"context"
	"fmt"
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/getter"
	"sigs.k8s.io/kustomize/api/internal/git"
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

	root, errD := demandDirectoryRoot(fSys, target)
	if errD == nil {
		return newLoaderAtConfirmedDir(lr, root, fSys, nil, getRepo), nil
	}

	ldr, errL := newLoaderAtGitClone(
		(&git.RepoSpec{}).WithRaw(target), fSys, nil, getRepo)

	if errL != nil {
		return nil, fmt.Errorf("Error demand directory %q and create loader %q", errD, errL)
	}

	return ldr, nil
}

func getRepo(repoSpec *git.RepoSpec) error {
	var err error
	repoSpec.Dir, err = filesys.NewTmpConfirmedDir()
	if err != nil {
		return err
	}

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	}

	opts := []getter.ClientOption{}
	client := &getter.Client{
		Ctx:     context.TODO(),
		Src:     repoSpec.Raw(),
		Dst:     repoSpec.Dir.String(),
		Pwd:     pwd,
		Mode:    getter.ClientModeAny,
		Options: opts,
	}
	return client.Get()
}
