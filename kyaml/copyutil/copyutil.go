// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The copyutil package contains libraries for copying directories of configuration.
package copyutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

// CopyDir copies a src directory to a dst directory.  CopyDir skips copying the .git directory from the src.
func CopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// don't copy the .git dir
		if path != src {
			rel := strings.TrimPrefix(path, src)
			if IsDotGitFolder(rel) {
				return nil
			}
		}

		// path is an absolute path, rather than a path relative to src.
		// e.g. if src is /path/to/package, then path might be /path/to/package/and/sub/dir
		// we need the path relative to src `and/sub/dir` when we are copying the files to dest.
		copyTo := strings.TrimPrefix(path, src)

		// make directories that don't exist
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, copyTo), info.Mode())
		}

		// copy file by reading and writing it
		b, err := ioutil.ReadFile(filepath.Join(src, copyTo))
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(dst, copyTo), b, info.Mode())
		if err != nil {
			return err
		}

		return nil
	})
}

// Diff returns a list of files that differ between the source and destination.
//
// Diff is guaranteed to return a non-empty set if any files differ, but
// this set is not guaranteed to contain all differing files.
func Diff(sourceDir, destDir string) (sets.String, error) {
	// get set of filenames in the package source
	upstreamFiles := sets.String{}
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip git repo if it exists
		if IsDotGitFolder(path) {
			return nil
		}

		upstreamFiles.Insert(strings.TrimPrefix(strings.TrimPrefix(path, sourceDir), string(filepath.Separator)))
		return nil
	})
	if err != nil {
		return sets.String{}, err
	}

	// get set of filenames in the cloned package
	localFiles := sets.String{}
	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip git repo if it exists
		if IsDotGitFolder(path) {
			return nil
		}

		localFiles.Insert(strings.TrimPrefix(strings.TrimPrefix(path, destDir), string(filepath.Separator)))
		return nil
	})
	if err != nil {
		return sets.String{}, err
	}

	// verify the source and cloned packages have the same set of filenames
	diff := upstreamFiles.SymmetricDifference(localFiles)

	// verify file contents match
	for _, f := range upstreamFiles.Intersection(localFiles).List() {
		fi, err := os.Stat(filepath.Join(destDir, f))
		if err != nil {
			return diff, err
		}
		if fi.Mode().IsDir() {
			// already checked that this directory exists in the local files
			continue
		}

		// compare upstreamFiles
		b1, err := ioutil.ReadFile(filepath.Join(destDir, f))
		if err != nil {
			return diff, err
		}
		b2, err := ioutil.ReadFile(filepath.Join(sourceDir, f))
		if err != nil {
			return diff, err
		}
		if !bytes.Equal(b1, b2) {
			fmt.Println(PrettyFileDiff(string(b1), string(b2)))
			diff.Insert(f)
		}
	}
	// return the differing files
	return diff, nil
}

// IsDotGitFolder checks if the provided path is either the .git folder or
// a file underneath the .git folder.
func IsDotGitFolder(path string) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(path))
	for _, c := range strings.Split(cleanPath, "/") {
		if c == ".git" {
			return true
		}
	}
	return false
}

// PrettyFileDiff takes the content of two files and returns the pretty diff
func PrettyFileDiff(s1, s2 string) string {
	dmp := diffmatchpatch.New()
	wSrc, wDst, warray := dmp.DiffLinesToRunes(s1, s2)
	diffs := dmp.DiffMainRunes(wSrc, wDst, false)
	diffs = dmp.DiffCharsToLines(diffs, warray)
	return dmp.DiffPrettyText(diffs)
}

// SyncFile copies file from src file path to a dst file path by replacement
// deletes dst file if src file doesn't exist
func SyncFile(src, dst string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		// delete dst if source doesn't exist
		if err = deleteFile(dst); err != nil {
			return err
		}
		return nil
	}

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	var filePerm os.FileMode

	// get the destination file perm if file exists
	dstFileInfo, err := os.Stat(dst)
	if err != nil {
		// get source file perm if destination file doesn't exist
		filePerm = srcFileInfo.Mode().Perm()
	} else {
		filePerm = dstFileInfo.Mode().Perm()
	}

	err = ioutil.WriteFile(dst, input, filePerm)
	if err != nil {
		return err
	}

	return nil
}

// deleteFile deletes file from path, returns no error if file doesn't exist
func deleteFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		// return nil if file doesn't exist
		return nil
	}
	return os.Remove(path)
}
