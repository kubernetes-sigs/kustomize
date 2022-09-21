// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/errors"
)

const (
	GeneratorWarning   = "generator extensions not yet supported in alpha"
	TransformerWarning = "transformer extensions not yet supported in alpha"
)

var (
	ErrInvalidRoot       = errors.Errorf("invalid root reference")
	ErrLocalizeDirExists = errors.Errorf("'%s' localize directory already exists", LocalizeDir)
	ErrNoRef             = errors.Errorf("localize remote root missing ref query string parameter")
)

// prefixRelErrWhenContains returns a prefix for the error in the event that filepath.Rel(basePath, targPath) returns one,
// where basePath contains targPath
func prefixRelErrWhenContains(basePath string, targPath string) string {
	return fmt.Sprintf("cannot find path from directory '%s' to '%s' inside directory", basePath, targPath)
}
