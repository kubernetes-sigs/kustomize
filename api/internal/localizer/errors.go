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
