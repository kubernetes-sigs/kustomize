// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var filesysBuilders = map[string]func() FileSystem{
	"MakeFsInMemory":       MakeFsInMemory,
	"MakeFsOnDisk":         MakeFsOnDisk,
	"MakeEmptyDirInMemory": func() FileSystem { return MakeEmptyDirInMemory() },
}

func TestNotExistErr(t *testing.T) {
	for name, builder := range filesysBuilders {
		t.Run(name, func(t *testing.T) {
			testNotExistErr(t, builder())
		})
	}
}

func testNotExistErr(t *testing.T, fs FileSystem) {
	const path = "bad-dir/file.txt"

	err := fs.RemoveAll(path)
	assert.Falsef(t, errors.Is(err, os.ErrNotExist), "RemoveAll should not return ErrNotExist, got %v", err)
	_, err = fs.Open(path)
	assert.Truef(t, errors.Is(err, os.ErrNotExist), "Open should return ErrNotExist, got %v", err)
	_, err = fs.ReadDir(path)
	assert.Truef(t, errors.Is(err, os.ErrNotExist), "ReadDir should return ErrNotExist, got %v", err)
	_, _, err = fs.CleanedAbs(path)
	assert.Truef(t, errors.Is(err, os.ErrNotExist), "CleanedAbs should return ErrNotExist, got %v", err)
	_, err = fs.ReadFile(path)
	assert.Truef(t, errors.Is(err, os.ErrNotExist), "ReadFile should return ErrNotExist, got %v", err)
	err = fs.Walk(path, func(_ string, _ os.FileInfo, err error) error { return err })
	assert.Truef(t, errors.Is(err, os.ErrNotExist), "Walk should return ErrNotExist, got %v", err)
}
