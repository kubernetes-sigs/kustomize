// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

const dirMsg = "expected '%s' to be a dir"

func makeTestDir(t *testing.T) (FileSystem, string) {
	t.Helper()
	req := require.New(t)

	fSys := MakeFsOnDisk()
	td := t.TempDir()

	testDir, err := filepath.EvalSymlinks(td)
	req.NoError(err)
	req.Truef(fSys.Exists(testDir), existMsg, testDir)
	req.Truef(fSys.IsDir(testDir), dirMsg, testDir)

	return fSys, testDir
}

func cleanWd(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	cleanedWd, err := filepath.EvalSymlinks(wd)
	require.NoError(t, err)

	return cleanedWd
}

func TestCleanedAbs_1(t *testing.T) {
	req := require.New(t)
	fSys, _ := makeTestDir(t)

	d, f, err := fSys.CleanedAbs("")
	req.NoError(err)

	wd := cleanWd(t)
	req.Equal(wd, d.String())
	req.Empty(f)
}

func TestCleanedAbs_2(t *testing.T) {
	req := require.New(t)
	fSys, _ := makeTestDir(t)

	root := getOSRoot(t)
	d, f, err := fSys.CleanedAbs(root)
	req.NoError(err)
	req.Equal(root, d.String())
	req.Empty(f)
}

func TestCleanedAbs_3(t *testing.T) {
	req := require.New(t)
	fSys, testDir := makeTestDir(t)

	err := fSys.WriteFile(
		filepath.Join(testDir, "foo"), []byte(`foo`))
	req.NoError(err)

	d, f, err := fSys.CleanedAbs(filepath.Join(testDir, "foo"))
	req.NoError(err)
	req.Equal(testDir, d.String())
	req.Equal("foo", f)
}

func TestCleanedAbs_4(t *testing.T) {
	req := require.New(t)
	fSys, testDir := makeTestDir(t)

	err := fSys.MkdirAll(filepath.Join(testDir, "d1", "d2"))
	req.NoError(err)

	err = fSys.WriteFile(
		filepath.Join(testDir, "d1", "d2", "bar"),
		[]byte(`bar`))
	req.NoError(err)

	d, f, err := fSys.CleanedAbs(
		filepath.Join(testDir, "d1", "d2"))
	req.NoError(err)
	req.Equal(filepath.Join(testDir, "d1", "d2"), d.String())
	req.Empty(f)

	d, f, err = fSys.CleanedAbs(
		filepath.Join(testDir, "d1", "d2", "bar"))
	req.NoError(err)
	req.Equal(filepath.Join(testDir, "d1", "d2"), d.String())
	req.Equal("bar", f)
}

func TestConfirmDirDisk(t *testing.T) {
	req := require.New(t)
	fSys, testDir := makeTestDir(t)
	wd := cleanWd(t)

	relDir := "actual_foo_431432"
	err := fSys.Mkdir(relDir)
	t.Cleanup(func() {
		err := fSys.RemoveAll(relDir)
		req.NoError(err)
	})
	req.NoError(err)
	req.Truef(fSys.Exists(relDir), existMsg, relDir)

	linkDir := filepath.Join(testDir, "pointer")
	err = os.Symlink(filepath.Join(wd, relDir), linkDir)
	req.NoError(err)

	root := getOSRoot(t)
	tests := map[string]*struct {
		path     string
		expected string
	}{
		"root": {
			root,
			root,
		},
		"non-selfdir relative path": {
			relDir,
			filepath.Join(wd, relDir),
		},
		"symlink": {
			linkDir,
			filepath.Join(wd, relDir),
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			actualPath, err := ConfirmDir(fSys, test.path)
			require.NoError(t, err)
			require.Equal(t, test.expected, actualPath.String())
		})
	}
}

func TestReadFilesRealFS(t *testing.T) {
	req := require.New(t)

	fSys, testDir := makeTestDir(t)

	dir := filepath.Join(testDir, "dir")
	nestedDir := filepath.Join(dir, "nestedDir")
	hiddenDir := filepath.Join(testDir, ".hiddenDir")
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
	req.NoError(err)

	err = fSys.MkdirAll(hiddenDir)
	req.NoError(err)

	// adding all files in every directory that we had defined
	for _, d := range dirs {
		req.Truef(fSys.IsDir(d), dirMsg, d)
		for _, f := range files {
			fPath := filepath.Join(d, f)
			err = fSys.WriteFile(fPath, []byte(f))
			req.NoError(err)
			req.Truef(fSys.Exists(fPath), existMsg, fPath)
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
				testDir: {dir},
				dir:     {nestedDir},
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
				testDir: {hiddenDir},
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
					expectedPaths = append(expectedPaths, filepath.Join(d, f))
				}
				if c.expectedDirs != nil {
					expectedPaths = append(expectedPaths, c.expectedDirs[d]...)
				}
				actualPaths, globErr := fSys.Glob(filepath.Join(d, c.globPattern))
				require.NoError(t, globErr)

				sort.Strings(actualPaths)
				sort.Strings(expectedPaths)
				if !reflect.DeepEqual(actualPaths, expectedPaths) {
					t.Fatalf("incorrect files found by glob: expected=%v, actual=%v", expectedPaths, actualPaths)
				}
			}
		})
	}
}
