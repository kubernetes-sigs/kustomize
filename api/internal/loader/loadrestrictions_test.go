// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/filesys"
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
	root := filesys.ConfirmedDir(
		filesys.Separator + filepath.Join("tmp", "foo"))
	path := filepath.Join(string(root), "whatever", "beans")

	fSys.Create(path)
	p, err := RestrictionRootOnly(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}

	// Legal.
	path = filepath.Join(
		string(root), "whatever", "..", "..", "foo", "whatever", "beans")
	p, err = RestrictionRootOnly(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(
		string(root), "whatever", "beans")
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}

	// Illegal; file exists but is out of bounds.
	path = filepath.Join(filesys.Separator+"tmp", "illegal")
	fSys.Create(path)
	_, err = RestrictionRootOnly(fSys, root, path)
	if err == nil {
		t.Fatal("should have an error")
	}
	// Normalize paths to forward slashes for cross-platform comparison
	expectedErr := fmt.Sprintf("file '%s' is not in or below '%s'",
		filepath.ToSlash(filepath.Join(filesys.Separator+"tmp", "illegal")),
		filepath.ToSlash(filepath.Join(filesys.Separator+"tmp", "foo")))
	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("unexpected err: %s", err)
	}
}
