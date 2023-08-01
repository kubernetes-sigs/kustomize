// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Run executes `kustomize localize` on fSys given the `localize` arguments and
// returns the path to the created newDir.
func Run(fSys filesys.FileSystem, target, scope, newDir string) (string, error) {
	dst, err := localizer.Run(target, scope, newDir, fSys)
	return dst, errors.Wrap(err)
}

// Pull executes `kustomize localize` on OCI artifacts
// returns the path to the created destination
func Pull(target, destination string) (string, error) {
	if destination == "" {
		destination = filesys.SelfDir
	}
	ociSpec, err := oci.NewOCISpecFromURL(target)
	if err != nil {
		return "", fmt.Errorf("[NewOCISepcFromURL] error: %w", err)
	}
	ociSpec.Dir = filesys.ConfirmedDir(destination)
	err = oci.PullArtifact(ociSpec)
	if err != nil {
		return "", fmt.Errorf("[PullArtifact] error: %w", err)
	}

	return destination, nil
}
