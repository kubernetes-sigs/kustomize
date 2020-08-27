// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package ext

// GetIgnoreFileName returns the name for ignore files in
// packages. It can be overridden by tools using this library.
var GetIgnoreFileName = func() string {
	return ".krmignore"
}
