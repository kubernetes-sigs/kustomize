// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// TODO: move most of the tests in the api/target package to this package.
package krusty_test

import (
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"testing"
)

// TODO: make this more like kusttest_test.AssertActualEqualsExpected
func assertOutput(t *testing.T, actual []byte, expected string) {
	if string(actual) != expected {
		t.Fatalf("Err: expected:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestSomething1(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	b := krusty.MakeKustomizer(fSys, krusty.MakeDefaultOptions())
	_, err := b.Run("hey")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "got file 'hey', but 'hey' must be a directory to be a root" {
		t.Fatalf("unexpected error: %v", err)
	}
}
