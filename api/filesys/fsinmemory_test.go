// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys_test

import (
	"bytes"
	"reflect"
	"testing"

	. "sigs.k8s.io/kustomize/api/filesys"
)

func TestExists(t *testing.T) {
	fSys := MakeFsInMemory()
	if fSys.Exists("foo") {
		t.Fatalf("expected no foo")
	}
	fSys.Mkdir("/")
	if !fSys.IsDir("/") {
		t.Fatalf("expected dir at /")
	}
}

func TestIsDir(t *testing.T) {
	fSys := MakeFsInMemory()
	expectedName := "my-dir"
	err := fSys.Mkdir(expectedName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	shouldExist(t, fSys, expectedName)
	if !fSys.IsDir(expectedName) {
		t.Fatalf(expectedName + " should be a dir")
	}
}

func shouldExist(t *testing.T, fSys FileSystem, name string) {
	if !fSys.Exists(name) {
		t.Fatalf(name + " should exist")
	}
}

func shouldNotExist(t *testing.T, fSys FileSystem, name string) {
	if fSys.Exists(name) {
		t.Fatalf(name + " should not exist")
	}
}

func TestRemoveAll(t *testing.T) {
	fSys := MakeFsInMemory()
	fSys.WriteFile("/foo/project/file.yaml", []byte("Unused"))
	fSys.WriteFile("/foo/project/subdir/file.yaml", []byte("Unused"))
	fSys.WriteFile("/foo/apple/subdir/file.yaml", []byte("Unused"))
	shouldExist(t, fSys, "/foo/project/file.yaml")
	shouldExist(t, fSys, "/foo/project/subdir/file.yaml")
	shouldExist(t, fSys, "/foo/apple/subdir/file.yaml")
	err := fSys.RemoveAll("/foo/project")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	shouldNotExist(t, fSys, "/foo/project/file.yaml")
	shouldNotExist(t, fSys, "/foo/project/subdir/file.yaml")
	shouldExist(t, fSys, "/foo/apple/subdir/file.yaml")
}

func TestIsDirDeeper(t *testing.T) {
	fSys := MakeFsInMemory()
	fSys.WriteFile("/foo/project/file.yaml", []byte("Unused"))
	fSys.WriteFile("/foo/project/subdir/file.yaml", []byte("Unused"))
	if !fSys.IsDir("/") {
		t.Fatalf("/ should be a dir")
	}
	if !fSys.IsDir("/foo") {
		t.Fatalf("/foo should be a dir")
	}
	if !fSys.IsDir("/foo/project") {
		t.Fatalf("/foo/project should be a dir")
	}
	if fSys.IsDir("/fo") {
		t.Fatalf("/fo should not be a dir")
	}
	if fSys.IsDir("/x") {
		t.Fatalf("/x should not be a dir")
	}
}

func TestCreate(t *testing.T) {
	fSys := MakeFsInMemory()
	f, err := fSys.Create("foo")
	if f == nil {
		t.Fatalf("expected file")
	}
	if err != nil {
		t.Fatalf("unexpected error")
	}
	shouldExist(t, fSys, "foo")
}

func TestReadFile(t *testing.T) {
	fSys := MakeFsInMemory()
	f, err := fSys.Create("foo")
	if f == nil {
		t.Fatalf("expected file")
	}
	if err != nil {
		t.Fatalf("unexpected error")
	}
	content, err := fSys.ReadFile("foo")
	if len(content) != 0 {
		t.Fatalf("expected no content")
	}
	if err != nil {
		t.Fatalf("expected no error")
	}
}

func TestWriteFile(t *testing.T) {
	fSys := MakeFsInMemory()
	c := []byte("heybuddy")
	err := fSys.WriteFile("foo", c)
	if err != nil {
		t.Fatalf("expected no error")
	}
	content, err := fSys.ReadFile("foo")
	if err != nil {
		t.Fatalf("expected read to work: %v", err)
	}
	if bytes.Compare(c, content) != 0 {
		t.Fatalf("incorrect content: %v", content)
	}
}

func TestGlob(t *testing.T) {
	fSys := MakeFsInMemory()
	fSys.Create("dir/foo")
	fSys.Create("dir/bar")
	files, err := fSys.Glob("dir/*")
	if err != nil {
		t.Fatalf("expected no error")
	}
	expected := []string{
		"dir/bar",
		"dir/foo",
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("incorrect files found by glob: %v", files)
	}
}
