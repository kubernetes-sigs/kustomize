/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package loader

import (
	"fmt"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

// Loader abstracts how bytes are read for manifest, resource, patch, or other
// files. Each Loader is tightly coupled with the location of a manifest file.
// So each file without an absolute path is located relative to the
// manifest file it was read from. Each Load() call is relative to this manifest
// location (referenced as root). The Loader hides how to read bytes from different
// "schemes" (e.g. file, url, or git).
type Loader interface {
	// Clones new Loader for a new app/package with its own manifest from the current
	// Loader. The "newRoot" can be relative or absolute. If it's relative, the new
	// Loader root is calculated from the current Loader root. Can be a file or directory.
	// If it's a file, then the base directory is used for root calculation.
	New(newRoot string) (Loader, error)
	// Returns the bytes at location or an error. If it's a relative path, then
	// the location is expanded using the Loader root.
	// Example: returns YAML bytes at location "/home/seans/project/service.yaml".
	Load(location string) ([]byte, error)
}

// Private implmentation of Loader interface.
type loaderImpl struct {
	root string
	fs   fs.FileSystem
	// http client for URL loading
	// git client for Git loading
}

// RootLoader initializes the first Loader, with the initial root location.
func RootLoader(root string, fs fs.FileSystem) Loader {
	// TODO: Validate the root
	return &loaderImpl{root: root, fs: fs}
}

// New clones a new Loader with a new absolute root path.
func (l *loaderImpl) New(newRoot string) (Loader, error) {
	loader, err := l.getSchemeLoader(newRoot)
	if err != nil {
		return nil, err
	}
	return &loaderImpl{root: loader.fullLocation(l.root, newRoot), fs: l.fs}, nil
}

// Load returns the bytes at the specified location.
// Implemented by getting a scheme-specific structure to
// load the bytes.
func (l *loaderImpl) Load(location string) ([]byte, error) {
	loader, err := l.getSchemeLoader(location)
	if err != nil {
		return nil, err
	}
	fullLocation := loader.fullLocation(l.root, location)
	return loader.load(fullLocation)
}

// Helper function to parse scheme from location parameter and return
func (l *loaderImpl) getSchemeLoader(location string) (schemeLoader, error) {
	// FIXME: First check the scheme of root location.
	switch {
	case isFilePath(location):
		return newFileLoader(l.fs)
	default:
		return nil, fmt.Errorf("unknown scheme: %v", location)
	}
}

// Parses the location to determine if it is a file path.
func isFilePath(location string) bool {
	return true
}

/////////////////////////////////////////////////
// Internal interface for specific type of loader
// Examples: fileLoader, HttpLoader, or GitLoader
type schemeLoader interface {
	// Combines the root and path into a full location string.
	fullLocation(root string, path string) string
	// Must be a full, non-relative location string.
	load(location string) ([]byte, error)
}
