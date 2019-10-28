// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
)

func TestRestrictionNone(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	root := filesys.ConfirmedDir("irrelevant")
	path := "whatever"
	p, err := RestrictionNone(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}
}

func TestRestrictionRootOnly(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	root := filesys.ConfirmedDir("/tmp/foo")

	path := "/tmp/foo/whatever/beans"
	p, err := RestrictionRootOnly(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}

	// Legal.
	path = "/tmp/foo/whatever/../../foo/whatever"
	p, err = RestrictionRootOnly(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	path = "/tmp/foo/whatever"
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}

	// Illegal.
	path = "/tmp/illegal"
	_, err = RestrictionRootOnly(fSys, root, path)
	if err == nil {
		t.Fatal("should have an error")
	}
	if !strings.Contains(
		err.Error(),
		"file '/tmp/illegal' is not in or below '/tmp/foo'") {
		t.Fatalf("unexpected err: %s", err)
	}
}
