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

	existMsg = "expected '%s' to exist"
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
func setupFileSys(t *testing.T, name string) (FileSystem, string, string) {
	t.Helper()

	switch name {
	case disk:
		fSys, testDir := makeTestDir(t)
		return fSys, testDir, cleanWd(t)
	case memoryAbs:
		return filesysBuilders[name](), Separator, Separator
	case memoryEmpty:
		return filesysBuilders[name](), "", ""
	default:
		t.Fatalf("unexpected FileSystem implementation '%s'", name)
		panic("unreachable point of execution")
	}
}

func TestConfirmDir(t *testing.T) {
	for name := range filesysBuilders {
		name := name
		t.Run(name, func(t *testing.T) {
			fSys, prefixPath, wd := setupFileSys(t, name)

			d1Path := filepath.Join(prefixPath, "d1")
			d2Path := filepath.Join(d1Path, ".d2")
			err := fSys.MkdirAll(d2Path)
			require.NoError(t, err)
			require.Truef(t, fSys.Exists(d2Path), existMsg, d2Path)

			tests := map[string]*struct {
				dir      string
				expected string
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
				test := test
				t.Run(subName, func(t *testing.T) {
					cleanDir, err := ConfirmDir(fSys, test.dir)
					require.NoError(t, err)
					require.Equal(t, test.expected, cleanDir.String())
				})
			}
		})
	}
}

func TestConfirmDirErr(t *testing.T) {
	for name := range filesysBuilders {
		thisName := name
		t.Run(thisName, func(t *testing.T) {
			fSys, prefixPath, _ := setupFileSys(t, thisName)

			fPath := filepath.Join(prefixPath, "foo")
			err := fSys.WriteFile(fPath, []byte(`foo`))
			require.NoError(t, err)
			require.Truef(t, fSys.Exists(fPath), existMsg, fPath)

			tests := map[string]string{
				"Empty":        "",
				"File":         fPath,
				"Non-existent": filepath.Join(prefixPath, "bar"),
			}
			for subName, invalidPath := range tests {
				invalidPath := invalidPath
				t.Run(subName, func(t *testing.T) {
					_, err := ConfirmDir(fSys, invalidPath)
					require.Error(t, err)
				})
			}
		})
	}
}
