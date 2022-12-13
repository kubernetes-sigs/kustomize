// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import "fmt"

type ResourceLoadError struct {
	InlineError error
	FileError   error
}

var _ error = ResourceLoadError{}

func (rle ResourceLoadError) Error() string {
	return fmt.Sprintf(`when parsing as inline received error: %s
when parsing as filepath received error: %s`, rle.InlineError, rle.FileError)
}

type PathLocalizeError struct {
	FileError error
	RootError error
}

var _ error = PathLocalizeError{}

func (ple PathLocalizeError) Error() string {
	return fmt.Sprintf(`when localizing as file received error: %s
when localizing as directory received error: %s`, ple.FileError, ple.RootError)
}
