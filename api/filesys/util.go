// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import "path/filepath"

// RootedPath returns a rooted path, e.g. "/foo/bar" as
// opposed to "foo/bar".
func RootedPath(elem ...string) string {
	return Separator + filepath.Join(elem...)
}

// StripTrailingSeps trims trailing filepath separators from input.
func StripTrailingSeps(s string) string {
	k := len(s)
	for k > 0 && s[k-1] == filepath.Separator {
		k--
	}
	return s[:k]
}

// StripLeadingSeps trims leading filepath separators from input.
func StripLeadingSeps(s string) string {
	k := 0
	for k < len(s) && s[k] == filepath.Separator {
		k++
	}
	return s[k:]
}
