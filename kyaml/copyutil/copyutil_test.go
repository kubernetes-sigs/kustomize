// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package copyutil_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/copyutil"
)

// TestDiff_identical verifies identical directories return an empty set
func TestDiff_identical(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(d, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.Empty(t, diff.List())
}

// TestDiff_additionalSourceFiles verifies if there are additional files
// in the source, the diff will contain them
func TestDiff_additionalSourceFiles(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a2"), 0700)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a1"), 0700)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.ElementsMatch(t, diff.List(), []string{"a2"})
}

// TestDiff_additionalDestFiles verifies if there are additional files
// in the dest, the diff will contain them
func TestDiff_additionalDestFiles(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a1"), 0700)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a2"), 0700)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.ElementsMatch(t, diff.List(), []string{"a2"})
}

// TestDiff_srcDestContentsDiffer verifies if the file contents
// differ between the source and destination, the diff
// contains the differing files
func TestDiff_srcDestContentsDiffer(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(d, "a1", "f.yaml"), []byte(`b`), 0600)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.ElementsMatch(t, diff.List(), []string{
		fmt.Sprintf("a1%sf.yaml", string(filepath.Separator)),
	})
}

// TestDiff_srcDestContentsDifferInDirs verifies if identical files
// exist in different directories, they are included in the diff
func TestDiff_srcDestContentsDifferInDirs(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "b1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(d, "b1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.ElementsMatch(t, diff.List(), []string{
		"a1",
		fmt.Sprintf("a1%sf.yaml", string(filepath.Separator)),
		fmt.Sprintf("b1%sf.yaml", string(filepath.Separator)),
		"b1",
	})
}

// TestDiff_skipGitSrc verifies that .git directories in the source
// are not looked at
func TestDiff_skipGitSrc(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	// files that just happen to start with .git should not be ignored.
	err = ioutil.WriteFile(
		filepath.Join(s, ".gitlab-ci.yml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	// git should be ignored
	err = os.Mkdir(filepath.Join(s, ".git"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, ".git", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(d, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = ioutil.WriteFile(
		filepath.Join(d, ".gitlab-ci.yml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.Empty(t, diff.List())
}

// TestDiff_skipGitDest verifies that .git directories in the destination
// are not looked at
func TestDiff_skipGitDest(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	d, err := ioutil.TempDir("", "copyutildest")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(d, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	// git should be ignored
	err = os.Mkdir(filepath.Join(d, ".git"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(d, ".git", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	diff, err := Diff(s, d)
	assert.NoError(t, err)
	assert.Empty(t, diff.List())
}

// TestSyncFile tests if destination file is replaced by source file content
func TestSyncFile(t *testing.T) {
	d1, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	d2, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	f1Name := d1 + "/temp.txt"
	f2Name := d2 + "/temp.txt"
	err = ioutil.WriteFile(f1Name, []byte("abc"), 0600)
	assert.NoError(t, err)
	err = ioutil.WriteFile(f2Name, []byte("def"), 0644)
	expectedFileInfo, _ := os.Stat(f2Name)
	assert.NoError(t, err)
	err = SyncFile(f1Name, f2Name)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(f2Name)
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(actual))
	dstFileInfo, _ := os.Stat(f2Name)
	assert.Equal(t, expectedFileInfo.Mode().String(), dstFileInfo.Mode().String())
}

// TestSyncFileNoDestFile tests if new file is created at destination with source file content
func TestSyncFileNoDestFile(t *testing.T) {
	d1, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	d2, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	f1Name := d1 + "/temp.txt"
	f2Name := d2 + "/temp.txt"
	err = ioutil.WriteFile(f1Name, []byte("abc"), 0644)
	assert.NoError(t, err)
	err = SyncFile(f1Name, f2Name)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(f2Name)
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(actual))
	dstFileInfo, _ := os.Stat(f2Name)
	srcFileInfo, _ := os.Stat(f1Name)
	assert.Equal(t, srcFileInfo.Mode().String(), dstFileInfo.Mode().String())
}

// TestSyncFileNoSrcFile tests if destination file is deleted if source file doesn't exist
func TestSyncFileNoSrcFile(t *testing.T) {
	d1, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	d2, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	f1Name := d1 + "/temp.txt"
	f2Name := d2 + "/temp.txt"
	err = ioutil.WriteFile(f2Name, []byte("abc"), 0644)
	assert.NoError(t, err)
	err = SyncFile(f1Name, f2Name)
	assert.NoError(t, err)
	_, err = ioutil.ReadFile(f2Name)
	assert.Error(t, err)
}

func TestPrettyFileDiff(t *testing.T) {
	s1 := `apiVersion: someversion/v1alpha2
kind: ContainerCluster
metadata:
  clusterName: "some_cluster"
  name: asm-cluster
  namespace: "PROJECT_ID" # {"$ref":"#/definitions/io.k8s.cli.setters.gcloud.core.project"}`

	s2 := `apiVersion: someversion/v1alpha2
kind: ContainerCluster
metadata:
  clusterName: "some_cluster"
  name: asm-cluster
  namespace: "some_project" # {"$ref":"#/definitions/io.k8s.cli.setters.gcloud.core.project"}`

	expectedLine1 := `[31m  namespace: "PROJECT_ID" # {"$ref":"#/definitions/io.k8s.cli.setters.gcloud.core.project"}`
	expectedLine2 := `[32m  namespace: "some_project" # {"$ref":"#/definitions/io.k8s.cli.setters.gcloud.core.project"}`

	assert.Contains(t, PrettyFileDiff(s1, s2), expectedLine1)
	assert.Contains(t, PrettyFileDiff(s1, s2), expectedLine2)
}

func TestCopyDir(t *testing.T) {
	s, err := ioutil.TempDir("", "copyutilsrc")
	assert.NoError(t, err)
	v, err := ioutil.TempDir("", "copyutilvalidate")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(s, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	// files that just happen to start with .git should not be ignored.
	err = ioutil.WriteFile(
		filepath.Join(s, ".gitlab-ci.yml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	// git should be ignored
	err = os.Mkdir(filepath.Join(s, ".git"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(s, ".git", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(v, "a1"), 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(
		filepath.Join(v, "a1", "f.yaml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	err = ioutil.WriteFile(
		filepath.Join(v, ".gitlab-ci.yml"), []byte(`a`), 0600)
	assert.NoError(t, err)

	d, err := ioutil.TempDir("", "copyutildestination")
	assert.NoError(t, err)

	err = CopyDir(s, d)
	assert.NoError(t, err)

	diff, err := Diff(d, v)
	assert.NoError(t, err)
	assert.Empty(t, diff.List())
}

func TestIsDotGitFolder(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		isDotGitFolder bool
	}{
		{
			name:           ".git folder",
			path:           "/foo/bar/.git",
			isDotGitFolder: true,
		},
		{
			name:           "subfolder of .git folder",
			path:           "/foo/.git/bar/zoo",
			isDotGitFolder: true,
		},
		{
			name:           "subfolder of .gitignore folder",
			path:           "/foo/.gitignore/bar",
			isDotGitFolder: false,
		},
		{
			name:           ".gitignore file",
			path:           "foo/bar/.gitignore",
			isDotGitFolder: false,
		},
		{
			name:           ".gitlab-ci.yml under .git folder",
			path:           "/foo/.git/bar/.gitignore",
			isDotGitFolder: true,
		},
		{
			name:           "windows path with .git folder",
			path:           "c:/foo/.git/bar",
			isDotGitFolder: true,
		},
		{
			name:           "windows path with .gitignore file",
			path:           "d:/foo/bar/.gitignore",
			isDotGitFolder: false,
		},
	}

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.isDotGitFolder, IsDotGitFolder(test.path))
		})
	}
}
