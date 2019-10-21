// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	. "sigs.k8s.io/kustomize/api/filesys"
)

func makeTestDir(t *testing.T) (FileSystem, string) {
	fSys := MakeFsOnDisk()
	td, err := ioutil.TempDir("", "kustomize_testing_dir")
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	testDir, err := filepath.EvalSymlinks(td)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !fSys.Exists(testDir) {
		t.Fatalf("expected existence")
	}
	if !fSys.IsDir(testDir) {
		t.Fatalf("expected directory")
	}
	return fSys, testDir
}

func TestCleanedAbs_1(t *testing.T) {
	fSys, testDir := makeTestDir(t)
	defer os.RemoveAll(testDir)

	d, f, err := fSys.CleanedAbs("")
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	if d.String() != wd {
		t.Fatalf("unexpected d=%s", d)
	}
	if f != "" {
		t.Fatalf("unexpected f=%s", f)
	}
}

func TestCleanedAbs_2(t *testing.T) {
	fSys, testDir := makeTestDir(t)
	defer os.RemoveAll(testDir)

	d, f, err := fSys.CleanedAbs("/")
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	if d != "/" {
		t.Fatalf("unexpected d=%s", d)
	}
	if f != "" {
		t.Fatalf("unexpected f=%s", f)
	}
}

func TestCleanedAbs_3(t *testing.T) {
	fSys, testDir := makeTestDir(t)
	defer os.RemoveAll(testDir)

	err := fSys.WriteFile(
		filepath.Join(testDir, "foo"), []byte(`foo`))
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}

	d, f, err := fSys.CleanedAbs(filepath.Join(testDir, "foo"))
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	if d.String() != testDir {
		t.Fatalf("unexpected d=%s", d)
	}
	if f != "foo" {
		t.Fatalf("unexpected f=%s", f)
	}
}

func TestCleanedAbs_4(t *testing.T) {
	fSys, testDir := makeTestDir(t)
	defer os.RemoveAll(testDir)

	err := fSys.MkdirAll(filepath.Join(testDir, "d1", "d2"))
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	err = fSys.WriteFile(
		filepath.Join(testDir, "d1", "d2", "bar"),
		[]byte(`bar`))
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}

	d, f, err := fSys.CleanedAbs(
		filepath.Join(testDir, "d1", "d2"))
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	if d.String() != filepath.Join(testDir, "d1", "d2") {
		t.Fatalf("unexpected d=%s", d)
	}
	if f != "" {
		t.Fatalf("unexpected f=%s", f)
	}

	d, f, err = fSys.CleanedAbs(
		filepath.Join(testDir, "d1", "d2", "bar"))
	if err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	if d.String() != filepath.Join(testDir, "d1", "d2") {
		t.Fatalf("unexpected d=%s", d)
	}
	if f != "bar" {
		t.Fatalf("unexpected f=%s", f)
	}
}

func TestReadFilesRealFS(t *testing.T) {
	fSys, testDir := makeTestDir(t)
	defer os.RemoveAll(testDir)

	err := fSys.WriteFile(path.Join(testDir, "foo"), []byte(`foo`))
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !fSys.Exists(path.Join(testDir, "foo")) {
		t.Fatalf("expected foo")
	}
	if fSys.IsDir(path.Join(testDir, "foo")) {
		t.Fatalf("expected foo not to be a directory")
	}

	err = fSys.WriteFile(path.Join(testDir, "bar"), []byte(`bar`))
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	files, err := fSys.Glob(path.Join("testDir", "*"))
	expected := []string{
		path.Join(testDir, "bar"),
		path.Join(testDir, "foo"),
	}
	if err != nil {
		t.Fatalf("expected no error")
	}
	if reflect.DeepEqual(files, expected) {
		t.Fatalf("incorrect files found by glob: %v", files)
	}
}
