// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fs

import "path/filepath"

// RPath returns a rooted path, e.g. "/hey/foo" as
// opposed to "hey/foo".
func RPath(elem ...string) string {
	return separator + filepath.Join(elem...)
}
