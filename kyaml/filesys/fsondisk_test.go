// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

const dirMsg = "Expected '%s' to be a dir \n"

func makeTestDir(t *testing.T, req *require.Assertions) (FileSystem, string) {
	t.Helper()

	fSys := MakeFsOnDisk()
	td := t.TempDir()

	testDir, err := filepath.EvalSymlinks(td)
	req.NoError(err)
	req.Truef(fSys.Exists(testDir), existMsg, testDir)
	req.Truef(fSys.IsDir(testDir), dirMsg, testDir)

	return fSys, testDir
}

func cleanWd(req *require.Assertions, fSys FileSystem) ConfirmedDir {
	wd, err := os.Getwd()
	req.NoError(err)

	cleanedWd, f, err := fSys.CleanedAbs(wd)
	req.NoError(err)
	req.Empty(f)

	return cleanedWd
}

func TestCleanedAbs_1(t *testing.T) {
	req := require.New(t)

	fSys, _ := makeTestDir(t, req)

	d, f, err := fSys.CleanedAbs("")
	req.NoError(err)

	wd := cleanWd(req, fSys)
	req.Equal(wd, d)
	req.Empty(f)
}

func TestCleanedAbs_2(t *testing.T) {
	req := require.New(t)

	fSys, _ := makeTestDir(t, req)

	d, f, err := fSys.CleanedAbs("/")
	req.NoError(err)
	req.Equal(ConfirmedDir("/"), d)
	req.Empty(f)
}

func TestCleanedAbs_3(t *testing.T) {
	req := require.New(t)

	fSys, testDir := makeTestDir(t, req)

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

	fSys, testDir := makeTestDir(t, req)

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

func TestDemandDirDisk(t *testing.T) {
	req := require.New(t)
	fSys, testDir := makeTestDir(t, req)

	wd := cleanWd(req, fSys)

	relDir := "actual-foo-431432"
	err := fSys.Mkdir(relDir)
	req.NoError(err)
	req.Truef(fSys.Exists(relDir), existMsg, relDir)
	t.Cleanup(func() {
		err := fSys.RemoveAll(relDir)
		req.NoError(err)
	})

	linkDir := filepath.Join(testDir, "pointer")
	err = os.Symlink(wd.Join(relDir), linkDir)
	req.NoError(err)

	tests := map[string]*struct {
		path      string
		cleanPath string
	}{
		"root": {
			"/",
			"/",
		},
		"non-selfdir relative path": {
			relDir,
			wd.Join(relDir),
		},
		"symlink": {
			linkDir,
			wd.Join(relDir),
		},
	}

	for name, test := range tests {
		tCase := test
		t.Run(name, func(t *testing.T) {
			actualPath, err := DemandDir(fSys, tCase.path)
			require.NoError(t, err)
			require.Equal(t, ConfirmedDir(tCase.cleanPath), actualPath)
		})
	}
}

func TestReadFilesRealFS(t *testing.T) {
	req := require.New(t)

	fSys, testDir := makeTestDir(t, req)

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
	req.NoError(err)

	err = fSys.MkdirAll(hiddenDir)
	req.NoError(err)

	// adding all files in every directory that we had defined
	for _, d := range dirs {
		req.Truef(fSys.IsDir(d), dirMsg, d)
		for _, f := range files {
			fPath := path.Join(d, f)
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
					expectedPaths = append(expectedPaths, path.Join(d, f))
				}
				if c.expectedDirs != nil {
					expectedPaths = append(expectedPaths, c.expectedDirs[d]...)
				}
				actualPaths, globErr := fSys.Glob(path.Join(d, c.globPattern))
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
