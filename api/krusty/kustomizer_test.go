// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// TODO: move most of the tests in the api/target package to this package.
package krusty_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

// TODO: move most of the tests in api/internal/target
// to this package, as they are all high level tests and
// examples appropriate to this level and package.
// The following test isn't much more than a usage example;
// everything is actually tested down in api/internal/target.
func TestSomething(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	b := krusty.MakeKustomizer(fSys, krusty.MakeDefaultOptions())
	_, err := b.Run("noSuchThing")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "'noSuchThing' doesn't exist" {
		t.Fatalf("unexpected error: %v", err)
	}
}
