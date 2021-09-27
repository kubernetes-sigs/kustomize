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
	var tests = []struct {
		name          string
		fileName      string
		recurse       bool
		outFilesCount int
	}{
		{
			name:          "dirs-with-file-recurse",
			fileName:      "Krmfile",
			outFilesCount: 3,
			recurse:       true,
		},
		{
			name:          "dirs-with-non-existent-file-recurse",
			fileName:      "non-existent-file.txt",
			outFilesCount: 0,
			recurse:       true,
		},
		{
			name:          "dir-with-file-no-recurse",
			fileName:      "Krmfile",
			outFilesCount: 1,
			recurse:       false,
		},
	}

	dir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	err = createTestDirStructure(dir)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			res, err := DirsWithFile(dir, test.fileName, test.recurse)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t, test.outFilesCount, len(res)) {
				t.FailNow()
			}
		})
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
