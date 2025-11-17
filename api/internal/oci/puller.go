// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"

	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Puller is a function that can clone a git repo.
type Puller func(repoSpec *RepoSpec) error

// PullUsingOciManifest pulls/copies the oci manifest definition
func PullUsingOciManifest(repoSpec *RepoSpec) error {
	ctx := context.Background()

	src, err := oci.New(repoSpec.RepoPath)
	if err != nil {
		return err
	}
	dir, err := filesys.NewTmpConfirmedDir()
	if err != nil {
		return err
	}

	repoSpec.Dir = dir

	fs, err := file.New(dir.String())
	if err != nil {
		return err
	}
	defer fs.Close()

	reference := "latest"
	if repoSpec.Tag != "" {
		reference = repoSpec.Tag
	} else if repoSpec.Digest != "" {
		reference = repoSpec.Digest
	}

	desc, err := oras.Copy(ctx, src, reference, fs, "", oras.DefaultCopyOptions)
	if err != nil {
		return err
	} else if repoSpec.Digest != "" && repoSpec.Digest != desc.Digest.String() {
		return errors.Errorf("expected digest %s, but pulled artifact with digest %s", repoSpec.Digest, desc.Digest)
	}

	return nil
}

// DoNothingPuller returns a puller that only sets
// pullDir field in the repoSpec.  It's assumed that
// the pullDir is associated with some fake filesystem
// used in a test.
func DoNothingPuller(dir filesys.ConfirmedDir) Puller {
	return func(rs *RepoSpec) error {
		rs.Dir = dir
		return nil
	}
}
