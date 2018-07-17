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

// Package loader has a data loading interface and various implementations.
package loader

import (
	"fmt"

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

// Loader interface exposes methods to read bytes in a scheme-agnostic manner.
// The Loader encapsulating a root location to calculate where to read from.
type Loader interface {
	// Root returns the scheme-specific string representing the root location for this Loader.
	Root() string
	// New returns Loader located at newRoot.
	New(newRoot string) (Loader, error)
	// Load returns the bytes read from the location or an error.
	Load(location string) ([]byte, error)
	// GlobLoad returns the bytes read from a glob path or an error.
	GlobLoad(location string) (map[string][]byte, error)
	// GetFileSystem returns filesystem used by this loader.
	GetFileSystem() fs.FileSystem
}

// Private implementation of Loader interface.
type loaderImpl struct {
	root    string
	schemes []SchemeLoader
	fsys    fs.FileSystem
}

// SchemeLoader is the interface for different types of loaders (e.g. fileLoader, httpLoader, etc.)
type SchemeLoader interface {
	// Does this location correspond to this scheme.
	IsScheme(root string, location string) bool
	// Combines the root and path into a full location string.
	FullLocation(root string, path string) (string, error)
	// Load bytes at scheme-specific location or an error.
	Load(location string) ([]byte, error)
	// GlobLoad returns the bytes read from a glob path or an error.
	GlobLoad(location string) (map[string][]byte, error)
}

const emptyRoot = ""

// Init initializes the first loader with the supported schemes.
// Example schemes: fileLoader, httpLoader, gitLoader.
func Init(schemes []SchemeLoader, fs fs.FileSystem) Loader {
	return &loaderImpl{root: emptyRoot, schemes: schemes, fsys: fs}
}

// Root returns the scheme-specific root location for this Loader.
func (l *loaderImpl) Root() string {
	return l.root
}

// Returns a new Loader rooted at newRoot. "newRoot" MUST be
// a directory (not a file). The directory can have a trailing
// slash or not.
// Example: "/home/seans/project" or "/home/seans/project/"
// NOT "/home/seans/project/file.yaml".
func (l *loaderImpl) New(newRoot string) (Loader, error) {
	scheme, err := l.getSchemeLoader(newRoot)
	if err != nil {
		return nil, err
	}
	root, err := scheme.FullLocation(l.root, newRoot)
	if err != nil {
		return nil, err
	}
	return &loaderImpl{root: root, schemes: l.schemes, fsys: l.GetFileSystem()}, nil
}

// Load returns all the bytes read from scheme-specific location or an error.
// "location" can be an absolute path, or if relative, full location is
// calculated from the Root().
func (l *loaderImpl) Load(location string) ([]byte, error) {
	scheme, err := l.getSchemeLoader(location)
	if err != nil {
		return nil, err
	}
	fullLocation, err := scheme.FullLocation(l.root, location)
	if err != nil {
		return nil, err
	}
	return scheme.Load(fullLocation)
}

// GlobLoad returns a map from path to bytes read from scheme-specific location or an error.
// "location" can be an absolute path, or if relative, full location is
// calculated from the Root().
func (l *loaderImpl) GlobLoad(location string) (map[string][]byte, error) {
	scheme, err := l.getSchemeLoader(location)
	if err != nil {
		return nil, err
	}
	fullLocation, err := scheme.FullLocation(l.root, location)
	if err != nil {
		return nil, err
	}
	return scheme.GlobLoad(fullLocation)
}

// Helper function to parse scheme from location parameter.
func (l *loaderImpl) getSchemeLoader(location string) (SchemeLoader, error) {
	for _, scheme := range l.schemes {
		if scheme.IsScheme(l.root, location) {
			return scheme, nil
		}
	}
	return nil, fmt.Errorf("Unknown Scheme: %s, %s\n", l.root, location)
}

// GetFileSystem returns filesystem used by this loader.
func (l *loaderImpl) GetFileSystem() fs.FileSystem {
	return l.fsys
}
