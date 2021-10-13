// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ConfirmedDir is a clean, absolute, delinkified path
// that was confirmed to point to an existing directory.
type ConfirmedDir string

// NewTmpConfirmedDir returns a temporary dir, else error.
// The directory is cleaned, no symlinks, etc. so it's
// returned as a ConfirmedDir.
func NewTmpConfirmedDir() (ConfirmedDir, error) {
	n, err := ioutil.TempDir("", "kustomize-")
	if err != nil {
		return "", err
	}

	// In MacOs `ioutil.TempDir` creates a directory
	// with root in the `/var` folder, which is in turn
	// a symlinked path to `/private/var`.
	// Function `filepath.EvalSymlinks`is used to
	// resolve the real absolute path.
	deLinked, err := filepath.EvalSymlinks(n)
	return ConfirmedDir(deLinked), err
}

// simplifyAndHash transforms the given string in such a way to avoid filepath
// collisions. A single level of a filepaths cannot contain slashes, so those
// must be removed. Some filesystems do differentiate between lower and
// uppercase letters, so those must be normalized as well. The resulting
// strings could easily collide, such as "example/git-repo"
// ("example-git-repo") and "Example-Git/repo" (also "example-git-repo"). This
// function also appends a hash of the given string to eliminate this ambiguity
// and prevent filepath collisions.
func simplifyAndHash(s string) string {
	var result string
	// Convert the original string to lowercase, and keep any alphanumeric
	// characters. Replace anything else with a dash character.
	for _, char := range strings.ToLower(s) {
		switch {
		case '0' <= char && char <= '9':
			result += string(char)
		case 'a' <= char && char <= 'z':
			result += string(char)
		default:
			result += "-"
		}
	}

	// Trim any dash characters from the start or end.
	result = strings.Trim(result, "-")

	// Compute a hash of the original string, and append it as a suffix.
	// Why sha1? The algorithm choice doesn't matter as we are only attempting
	// to avoid filepath collisions. This is not considered a cryptographic
	// operation, so we do not need to prevent hash reversals.
	hash := sha1.Sum([]byte(s))
	result += "-" + hex.EncodeToString(hash[:])

	return result
}

// ErrCachedDirExists is a sentinel error indicating that a cached directory
// currently exists, and does not need to be repopulated by (for example)
// running a git clone.
var ErrCachedDirExists = errors.New("cached directory exists")

// NewCachedConfirmedDir returns a cache directory. This directory is rooted
// under the given cacheDir and named after the other given git values. The
// sentinel error ErrCachedDirExists is returned in the event that the cache
// directory already exists.
func NewCachedConfirmedDir(cacheDir, gitHost, gitRepo, gitRef string) (ConfirmedDir, error) {
	dir := filepath.Join(
		cacheDir,
		simplifyAndHash(gitHost),
		simplifyAndHash(gitRepo),
		gitRef,
	)

	// Check if the final cache dir exists and was cloned into. If it does, we
	// know that it was populated by a previous clone operation, which
	// subsequently can now be ignored. Return the ErrCachedDirExists sentinel
	// error to signal to the caller that the cache directory already exists.
	if _, err := os.Stat(filepath.Join(dir, ".git", "index")); err == nil {
		return ConfirmedDir(dir), ErrCachedDirExists
	}

	// There is a chance that the final cache directory was partially
	// initialized, for example if a user hit ^C, or took too long to unlock
	// their SSH key. Attempt to remove the directory, and if it doesn't exist
	// then it's just a nop.
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}

	// Create the final cache directory.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return ConfirmedDir(dir), nil
}

// HasPrefix returns true if the directory argument
// is a prefix of self (d) from the point of view of
// a file system.
//
// I.e., it's true if the argument equals or contains
// self (d) in a file path sense.
//
// HasPrefix emulates the semantics of strings.HasPrefix
// such that the following are true:
//
//   strings.HasPrefix("foobar", "foobar")
//   strings.HasPrefix("foobar", "foo")
//   strings.HasPrefix("foobar", "")
//
//   d := fSys.ConfirmDir("/foo/bar")
//   d.HasPrefix("/foo/bar")
//   d.HasPrefix("/foo")
//   d.HasPrefix("/")
//
// Not contacting a file system here to check for
// actual path existence.
//
// This is tested on linux, but will have trouble
// on other operating systems.
// TODO(monopole) Refactor when #golang/go/18358 closes.
// See also:
//   https://github.com/golang/go/issues/18358
//   https://github.com/golang/dep/issues/296
//   https://github.com/golang/dep/blob/master/internal/fs/fs.go#L33
//   https://codereview.appspot.com/5712045
func (d ConfirmedDir) HasPrefix(path ConfirmedDir) bool {
	if path.String() == string(filepath.Separator) || path == d {
		return true
	}
	return strings.HasPrefix(
		string(d),
		string(path)+string(filepath.Separator))
}

func (d ConfirmedDir) Join(path string) string {
	return filepath.Join(string(d), path)
}

func (d ConfirmedDir) String() string {
	return string(d)
}
