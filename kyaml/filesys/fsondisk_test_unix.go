// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package filesys

import "testing"

func getOSRoot(t *testing.T) string {
	t.Helper()
	return "/"
}
