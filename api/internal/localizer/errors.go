// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import "fmt"

type ResourceLoadError struct {
	InlineError error
	FileError   error
}

func (rle ResourceLoadError) Error() string {
	return fmt.Sprintf(`when parsing as inline received error: %s
when parsing as filepath received error: %s`, rle.InlineError, rle.FileError)
}

type PathLocalizeError struct {
	Path      string
	FileError error
	RootError error
}

func (ple PathLocalizeError) Error() string {
	return fmt.Sprintf(`could not localize path %q as file: %s; could not localize path %q as directory: %s`,
		ple.Path, ple.FileError, ple.Path, ple.RootError)
}
