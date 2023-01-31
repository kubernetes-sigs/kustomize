// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Run executes `kustomize localize` on fSys given the `localize` arguments and
// returns the path to the created newDir.
func Run(fSys filesys.FileSystem, target, scope, newDir string) (string, error) {
	dst, err := localizer.Run(target, scope, newDir, fSys)
	return dst, errors.Wrap(err)
}
