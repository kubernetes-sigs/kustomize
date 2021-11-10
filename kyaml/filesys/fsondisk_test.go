// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows
// +build !windows

package filesys

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
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

	dir := path.Join(testDir, "dir")
	nestedDir := path.Join(dir, "nestedDir")
	hiddenDir := path.Join(testDir, ".hiddenDir")
	dirs := []string{
		testDir,
		dir,
		nestedDir,
		hiddenDir,
	}
	// all directories will have all these files
	files := []string{
		"bar",
		"foo",
		"file-1.xtn",
		".file-2.xtn",
		".some-file-3.xtn",
		".some-file-4.xtn",
	}

	err := fSys.MkdirAll(nestedDir)
	if err != nil {
		t.Fatalf("Unexpected Error %v\n", err)
	}
	err = fSys.MkdirAll(hiddenDir)
	if err != nil {
		t.Fatalf("Unexpected Error %v\n", err)
	}

	// adding all files in every directory that we had defined
	for _, d := range dirs {
		if !fSys.IsDir(d) {
			t.Fatalf("Expected %s to be a dir\n", d)
		}
		for _, f := range files {
			err = fSys.WriteFile(path.Join(d, f), []byte(f))
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}
			if !fSys.Exists(path.Join(d, f)) {
				t.Fatalf("expected %s", f)
			}
		}
	}

	tests := map[string]struct {
		globPattern   string
		expectedFiles []string
		expectedDirs  map[string][]string // glob returns directories as well, so we need to add those to expected files
	}{
		"AllVisibleFiles": {
			globPattern: "*",
			expectedFiles: []string{
				"bar",
				"foo",
				"file-1.xtn",
			},
			expectedDirs: map[string][]string{
				testDir: []string{dir},
				dir:     []string{nestedDir},
			},
		},
		"AllHiddenFiles": {
			globPattern: ".*",
			expectedFiles: []string{
				".file-2.xtn",
				".some-file-3.xtn",
				".some-file-4.xtn",
			},
			expectedDirs: map[string][]string{
				testDir: []string{hiddenDir},
			},
		},
		"foo_File": {
			globPattern: "foo",
			expectedFiles: []string{
				"foo",
			},
		},
		"dotsome-file_PrefixedFiles": {
			globPattern: ".some-file*",
			expectedFiles: []string{
				".some-file-3.xtn",
				".some-file-4.xtn",
			},
		},
	}

	for n, c := range tests {
		t.Run(n, func(t *testing.T) {
			for _, d := range dirs {
				var expectedPaths []string
				for _, f := range c.expectedFiles {
					expectedPaths = append(expectedPaths, path.Join(d, f))
				}
				if c.expectedDirs != nil {
					expectedPaths = append(expectedPaths, c.expectedDirs[d]...)
				}
				actualPaths, globErr := fSys.Glob(path.Join(d, c.globPattern))
				if globErr != nil {
					t.Fatalf("Unexpected Error : %v\n", globErr)
				}
				sort.Strings(actualPaths)
				sort.Strings(expectedPaths)
				if !reflect.DeepEqual(actualPaths, expectedPaths) {
					t.Fatalf("incorrect files found by glob: expected=%v, actual=%v", expectedPaths, actualPaths)
				}
			}
		})
	}
}
