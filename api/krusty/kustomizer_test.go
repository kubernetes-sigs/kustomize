// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// A simple usage example to shows what happens when
// there are no files to read.
// For more substantial tests and examples,
// see other tests in this package.
func TestEmptyFileSystem(t *testing.T) {
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	_, _,err := b.Run(filesys.MakeFsInMemory(), "noSuchThing")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "'noSuchThing' doesn't exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}
