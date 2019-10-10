// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package fs provides a file system abstraction layer.
package fs

import (
	"io"
	"os"
	"path/filepath"
)

// FileSystem groups basic os filesystem methods.
type FileSystem interface {
	// Create a file.
	Create(name string) (File, error)
	// MkDir makes a directory.
	Mkdir(path string) error
	// MkDir makes a directory path, creating intervening directories.
	MkdirAll(path string) error
	// RemoveAll removes path and any children it contains.
	RemoveAll(path string) error
	// Open opens the named file for reading.
	Open(path string) (File, error)
	// IsDir returns true if the path is a directory.
	IsDir(path string) bool
	// CleanedAbs converts the given path into a
	// directory and a file name, where the directory
	// is represented as a ConfirmedDir and all that implies.
	// If the entire path is a directory, the file component
	// is an empty string.
	CleanedAbs(path string) (ConfirmedDir, string, error)
	// Exists is true if the path exists in the file system.
	Exists(path string) bool
	// Glob returns the list of matching files
	Glob(pattern string) ([]string, error)
	// ReadFile returns the contents of the file at the given path.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes the data to a file at the given path.
	WriteFile(path string, data []byte) error
	// Walk walks the file system with the given WalkFunc.
	Walk(path string, walkFn filepath.WalkFunc) error
}

// File groups the basic os.File methods.
type File interface {
	io.ReadWriteCloser
	Stat() (os.FileInfo, error)
}
