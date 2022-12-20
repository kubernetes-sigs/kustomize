// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizetests

import (
	"sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// LocalizeRunner runs `kustomize localize`. Like a CLI subprocess executing the command,
// it only needs the command arguments and flags and to be injected with a file system to run.
type LocalizeRunner struct {
	Scope string
}

// Run `localize`s referenced files in kustomization target to destination newDir on fSys.
func (lr *LocalizeRunner) Run(fSys filesys.FileSystem, target string, newDir string) error {
	return errors.Wrap(localizer.Run(target, lr.Scope, newDir, fSys))
}
