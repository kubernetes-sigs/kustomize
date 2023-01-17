// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizertest

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// PrepareFs returns an in-memory and the actual file system, both with the test
// directory with the file content mapping in files.
// dirs are the directory paths that need to be created to write the files.
func PrepareFs(t *testing.T, dirs []string, files map[string]string) (
	memory, actual filesys.FileSystem, test filesys.ConfirmedDir) {
	t.Helper()

	memory = filesys.MakeFsInMemory()
	actual = filesys.MakeFsOnDisk()

	testDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)

	SetupDir(t, memory, testDir.String(), files)
	for _, dirPath := range dirs {
		require.NoError(t, actual.MkdirAll(testDir.Join(dirPath)))
	}
	SetupDir(t, actual, testDir.String(), files)

	t.Cleanup(func() {
		_ = actual.RemoveAll(testDir.String())
	})

	return memory, actual, testDir
}

// SetupDir creates each file, specified by the file name to content mapping in
// files, under dir on fSys
func SetupDir(t *testing.T, fSys filesys.FileSystem, dir string,
	files map[string]string) {
	t.Helper()

	for file, content := range files {
		require.NoError(t, fSys.WriteFile(filepath.Join(dir, file), []byte(content)))
	}
}

// SetWorkingDir sets the working directory to workingDir and restores the
// original working directory after test completion.
func SetWorkingDir(t *testing.T, workingDir string) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	err = os.Chdir(workingDir)
	require.NoError(t, err)
}

// CheckFs checks actual, the real file system, against expected, a file
// system in memory, for contents in directory dir.
// CheckFs does not allow symlinks.
func CheckFs(t *testing.T, dir string, expected, actual filesys.FileSystem) {
	t.Helper()

	err := actual.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		require.NotEqual(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)
		require.True(t, expected.Exists(path), "unexpected file %q", path)
		return nil
	})
	require.NoError(t, err)

	err = expected.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		if info.IsDir() {
			require.DirExists(t, path)
		} else {
			require.FileExists(t, path)

			expectedContent, err := expected.ReadFile(path)
			require.NoError(t, err)
			actualContent, err := actual.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, string(expectedContent), string(actualContent))
		}
		return nil
	})
	require.NoError(t, err)
}
