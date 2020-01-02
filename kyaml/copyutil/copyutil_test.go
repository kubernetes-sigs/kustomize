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
