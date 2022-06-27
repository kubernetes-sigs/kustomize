// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows
// +build !windows

package filesys

import "testing"

func getOSRoot(t *testing.T) string {
	t.Helper()
	return "/"
}
