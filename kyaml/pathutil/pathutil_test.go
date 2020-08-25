// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubDirsWithFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	err = createTestDirStructure(dir)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	res, err := SubDirsWithFile(dir, "Krmfile")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, 3, len(res)) {
		t.FailNow()
	}
}

func TestSubDirsWithFileNoMatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	res, err := SubDirsWithFile(dir, "non-existent-file.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	var expected []string
	if !assert.Equal(t, expected, res) {
		t.FailNow()
	}
}

func createTestDirStructure(dir string) error {
	/*
		Adds the folders to the input dir with following structure
		dir
		├── Krmfile
		├── subpkg1
		│   ├── Krmfile
		│   └── subdir1
		└── subpkg2
		    └── Krmfile
	*/
	err := os.MkdirAll(filepath.Join(dir, "subpkg1/subdir1"), 0777|os.ModeDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(dir, "subpkg2"), 0777|os.ModeDir)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg1", "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg2", "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	return nil
}
