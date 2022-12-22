// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Run `kustomize localize`s files referenced by kustomization target in scope to destination newDir on fSys.
func Run(fSys filesys.FileSystem, target, scope, newDir string) error {
	return errors.Wrap(localizer.Run(target, scope, newDir, fSys))
}
