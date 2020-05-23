// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
)

func GoBin() string {
	return filepath.Join(runtime.GOROOT(), "bin", "go")
}

// DeterminePluginSrcRoot guesses where the user
// has her ${g}/${v}/$lower(${k})/${k}.go files.
func DeterminePluginSrcRoot(fSys filesys.FileSystem) (string, error) {
	return konfig.FirstDirThatExistsElseError(
		"source directory", fSys, []konfig.NotedFunc{
			{
				Note: "relative to unit test",
				F: func() string {
					return filepath.Clean(
						filepath.Join(
							os.Getenv("PWD"),
							"..", "..",
							konfig.RelPluginHome))
				},
			},
			{
				Note: "relative to unit test (internal pkg)",
				F: func() string {
					return filepath.Clean(
						filepath.Join(
							os.Getenv("PWD"),
							"..", "..", "..", "..",
							konfig.RelPluginHome))
				},
			},
			{
				Note: "relative to api package",
				F: func() string {
					return filepath.Clean(
						filepath.Join(
							os.Getenv("PWD"),
							"..", "..", "..",
							konfig.RelPluginHome))
				},
			},
			{
				Note: "old style $GOPATH",
				F: func() string {
					return filepath.Join(
						os.Getenv("GOPATH"),
						"src", konfig.DomainName,
						konfig.ProgramName, konfig.RelPluginHome)
				},
			},
			{
				Note: "HOME with literal 'gopath'",
				F: func() string {
					return filepath.Join(
						konfig.HomeDir(), "gopath",
						"src", konfig.DomainName,
						konfig.ProgramName, konfig.RelPluginHome)
				},
			},
			{
				Note: "home directory",
				F: func() string {
					return filepath.Join(
						konfig.HomeDir(), konfig.DomainName,
						konfig.ProgramName, konfig.RelPluginHome)
				},
			},
		})
}

// FileYoungerThan returns true if the file both exists and has an
// age is <= the Duration argument.
func FileYoungerThan(path string, d time.Duration) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return time.Since(fi.ModTime()) <= d
}

// FileModifiedAfter returns true if the file both exists and was
// modified after the given time..
func FileModifiedAfter(path string, t time.Time) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return fi.ModTime().After(t)
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
