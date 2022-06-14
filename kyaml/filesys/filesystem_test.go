// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	disk        = "MakeFsOnDisk"
	memoryAbs   = "MakeFsInMemory"
	memoryEmpty = "MakeEmptyDirInMemory"

	existMsg = "Expected '%s' to exist \n"
)

var filesysBuilders = map[string]func() FileSystem{
	memoryAbs:   MakeFsInMemory,
	disk:        MakeFsOnDisk,
	memoryEmpty: func() FileSystem { return MakeEmptyDirInMemory() },
}

func TestNotExistErr(t *testing.T) {
	for name, builder := range filesysBuilders {
		t.Run(name, func(t *testing.T) {
			testNotExistErr(t, builder())
		})
	}
}

func testNotExistErr(t *testing.T, fs FileSystem) {
	t.Helper()
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

// setupFileSys returns file system, default directory to test in, and working directory
func setupFileSys(t *testing.T, req *require.Assertions, name string, bldr func() FileSystem) (FileSystem, string, string) {
	t.Helper()
	switch name {
	case disk:
		fSys, testDir := makeTestDir(t, req)
		return fSys, testDir, string(cleanWd(req, fSys))
	case memoryAbs:
		return bldr(), Separator, Separator
	default:
		req.Equal(memoryEmpty, name)
		return bldr(), "", ""
	}
}

func TestDemandDir(t *testing.T) {
	for name, builder := range filesysBuilders {
		thisName := name
		thisBldr := builder
		t.Run(thisName, func(t *testing.T) {
			req := require.New(t)
			fSys, prefixPath, wd := setupFileSys(t, req, thisName, thisBldr)

			d1Path := filepath.Join(prefixPath, "d1")
			d2Path := filepath.Join(d1Path, ".d2")
			err := fSys.MkdirAll(d2Path)
			req.NoError(err)
			req.Truef(fSys.Exists(d2Path), existMsg, d2Path)

			tests := map[string]*struct {
				dir      string
				cleanDir string
			}{
				"Simple": {
					d1Path,
					d1Path,
				},
				"Hidden": {
					d2Path,
					d2Path,
				},
				"Relative": {
					SelfDir,
					wd,
				},
			}
			for subName, test := range tests {
				tCase := test
				t.Run(subName, func(t *testing.T) {
					cleanDir, err := DemandDir(fSys, tCase.dir)
					require.NoError(t, err)
					require.Equal(t, ConfirmedDir(tCase.cleanDir), cleanDir)
				})
			}
		})
	}
}

func TestDemandDirErr(t *testing.T) {
	for name, builder := range filesysBuilders {
		thisName := name
		thisBldr := builder
		t.Run(thisName, func(t *testing.T) {
			req := require.New(t)
			fSys, prefixPath, _ := setupFileSys(t, req, thisName, thisBldr)

			fPath := filepath.Join(prefixPath, "foo")
			err := fSys.WriteFile(fPath, []byte(`foo`))
			req.NoError(err)
			req.Truef(fSys.Exists(fPath), existMsg, fPath)

			tests := map[string]string{
				"Empty":          "",
				"File":           fPath,
				"Does not exist": filepath.Join(prefixPath, "bar"),
			}
			for subName, invalidPath := range tests {
				path := invalidPath
				t.Run(subName, func(t *testing.T) {
					_, err := DemandDir(fSys, path)
					require.Error(t, err)
				})
			}
		})
	}
}
