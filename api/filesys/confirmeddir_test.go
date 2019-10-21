// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys_test

import (
	"path/filepath"
	"testing"

	. "sigs.k8s.io/kustomize/api/filesys"
)

func TestJoin(t *testing.T) {
	fSys := MakeFsInMemory()
	err := fSys.Mkdir("/foo")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	d, f, err := fSys.CleanedAbs("/foo")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if f != "" {
		t.Fatalf("unexpected file: %v", f)
	}
	if d.Join("bar") != "/foo/bar" {
		t.Fatalf("expected join %s", d.Join("bar"))
	}
}

func TestHasPrefix_Slash(t *testing.T) {
	fSys := MakeFsInMemory()
	d, f, err := fSys.CleanedAbs("/")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if f != "" {
		t.Fatalf("unexpected file: %v", f)
	}
	if d.HasPrefix("/hey") {
		t.Fatalf("should be false")
	}
	if !d.HasPrefix("/") {
		t.Fatalf("/ should have the prefix /")
	}
}

func TestHasPrefix_SlashFoo(t *testing.T) {
	fSys := MakeFsInMemory()
	err := fSys.Mkdir("/foo")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	d, _, err := fSys.CleanedAbs("/foo")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if d.HasPrefix("/fo") {
		t.Fatalf("/foo does not have path prefix /fo")
	}
	if d.HasPrefix("/fod") {
		t.Fatalf("/foo does not have path prefix /fod")
	}
	if !d.HasPrefix("/foo") {
		t.Fatalf("/foo should have prefix /foo")
	}
}

func TestHasPrefix_SlashFooBar(t *testing.T) {
	fSys := MakeFsInMemory()
	err := fSys.MkdirAll("/foo/bar")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	d, _, err := fSys.CleanedAbs("/foo/bar")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if d.HasPrefix("/fo") {
		t.Fatalf("/foo/bar does not have path prefix /fo")
	}
	if d.HasPrefix("/foobar") {
		t.Fatalf("/foo/bar does not have path prefix /foobar")
	}
	if !d.HasPrefix("/foo/bar") {
		t.Fatalf("/foo/bar should have prefix /foo/bar")
	}
	if !d.HasPrefix("/foo") {
		t.Fatalf("/foo/bar should have prefix /foo")
	}
	if !d.HasPrefix("/") {
		t.Fatalf("/foo/bar should have prefix /")
	}
}

func TestNewTempConfirmDir(t *testing.T) {
	tmp, err := NewTmpConfirmedDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	delinked, err := filepath.EvalSymlinks(string(tmp))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(tmp) != delinked {
		t.Fatalf("unexpected path containing symlinks")
	}
}
