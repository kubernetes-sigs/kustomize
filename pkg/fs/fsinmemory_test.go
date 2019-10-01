// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"bytes"
	"reflect"
	"testing"
)

func TestExists(t *testing.T) {
	x := MakeFsInMemory()
	if x.Exists("foo") {
		t.Fatalf("expected no foo")
	}
	x.Mkdir("/")
	if !x.IsDir("/") {
		t.Fatalf("expected dir at /")
	}
}

func TestIsDir(t *testing.T) {
	x := MakeFsInMemory()
	expectedName := "my-dir"
	err := x.Mkdir(expectedName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	shouldExist(t, x, expectedName)
	if !x.IsDir(expectedName) {
		t.Fatalf(expectedName + " should be a dir")
	}
}

func shouldExist(t *testing.T, fs FileSystem, name string) {
	if !fs.Exists(name) {
		t.Fatalf(name + " should exist")
	}
}

func shouldNotExist(t *testing.T, fs FileSystem, name string) {
	if fs.Exists(name) {
		t.Fatalf(name + " should not exist")
	}
}

func TestRemoveAll(t *testing.T) {
	x := MakeFsInMemory()
	x.WriteFile("/foo/project/file.yaml", []byte("Unused"))
	x.WriteFile("/foo/project/subdir/file.yaml", []byte("Unused"))
	x.WriteFile("/foo/apple/subdir/file.yaml", []byte("Unused"))
	shouldExist(t, x, "/foo/project/file.yaml")
	shouldExist(t, x, "/foo/project/subdir/file.yaml")
	shouldExist(t, x, "/foo/apple/subdir/file.yaml")
	err := x.RemoveAll("/foo/project")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	shouldNotExist(t, x, "/foo/project/file.yaml")
	shouldNotExist(t, x, "/foo/project/subdir/file.yaml")
	shouldExist(t, x, "/foo/apple/subdir/file.yaml")
}

func TestIsDirDeeper(t *testing.T) {
	x := MakeFsInMemory()
	x.WriteFile("/foo/project/file.yaml", []byte("Unused"))
	x.WriteFile("/foo/project/subdir/file.yaml", []byte("Unused"))
	if !x.IsDir("/") {
		t.Fatalf("/ should be a dir")
	}
	if !x.IsDir("/foo") {
		t.Fatalf("/foo should be a dir")
	}
	if !x.IsDir("/foo/project") {
		t.Fatalf("/foo/project should be a dir")
	}
	if x.IsDir("/fo") {
		t.Fatalf("/fo should not be a dir")
	}
	if x.IsDir("/x") {
		t.Fatalf("/x should not be a dir")
	}
}

func TestCreate(t *testing.T) {
	x := MakeFsInMemory()
	f, err := x.Create("foo")
	if f == nil {
		t.Fatalf("expected file")
	}
	if err != nil {
		t.Fatalf("unexpected error")
	}
	shouldExist(t, x, "foo")
}

func TestReadFile(t *testing.T) {
	x := MakeFsInMemory()
	f, err := x.Create("foo")
	if f == nil {
		t.Fatalf("expected file")
	}
	if err != nil {
		t.Fatalf("unexpected error")
	}
	content, err := x.ReadFile("foo")
	if len(content) != 0 {
		t.Fatalf("expected no content")
	}
	if err != nil {
		t.Fatalf("expected no error")
	}
}

func TestWriteFile(t *testing.T) {
	x := MakeFsInMemory()
	c := []byte("heybuddy")
	err := x.WriteFile("foo", c)
	if err != nil {
		t.Fatalf("expected no error")
	}
	content, err := x.ReadFile("foo")
	if err != nil {
		t.Fatalf("expected read to work: %v", err)
	}
	if bytes.Compare(c, content) != 0 {
		t.Fatalf("incorrect content: %v", content)
	}
}

func TestGlob(t *testing.T) {
	x := MakeFsInMemory()
	x.Create("dir/foo")
	x.Create("dir/bar")
	files, err := x.Glob("dir/*")
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
