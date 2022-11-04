// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"
)

type InvalidRootError struct{}

func (ir InvalidRootError) Error() string {
	return "invalid root reference"
}

type LocalizeDirExistsError struct{}

func (lde LocalizeDirExistsError) Error() string {
	return LocalizeDir + " localize directory already exists"
}

type NoRefError struct {
	Root string
}

func (nr NoRefError) Error() string {
	return fmt.Sprintf("localize remote root %q missing ref query string parameter", nr.Root)
}

// prefixRelErrWhenContains returns a prefix for the error in the event that filepath.Rel(basePath, targPath) returns one,
// where basePath contains targPath
func prefixRelErrWhenContains(basePath string, targPath string) string {
	return fmt.Sprintf("cannot find path from directory %q to %q inside directory", basePath, targPath)
}
