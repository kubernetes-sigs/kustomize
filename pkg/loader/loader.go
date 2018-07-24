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

import "fmt"

// Loader interface exposes methods to read bytes.
type Loader interface {
	// Root returns the root location for this Loader.
	Root() string
	// New returns Loader located at newRoot.
	New(newRoot string) (Loader, error)
	// Load returns the bytes read from the location or an error.
	Load(location string) ([]byte, error)
	// GlobLoad returns the bytes read from a glob path or an error.
	GlobLoad(location string) (map[string][]byte, error)
}

// Private implementation of Loader interface.
type loaderImpl struct {
	root    string
	fLoader *FileLoader
}

const emptyRoot = ""

// NewLoader initializes the first loader with the supported fLoader.
func NewLoader(fl *FileLoader) Loader {
	return &loaderImpl{root: emptyRoot, fLoader: fl}
}

// Root returns the root location for this Loader.
func (l *loaderImpl) Root() string {
	return l.root
}

// Returns a new Loader rooted at newRoot. "newRoot" MUST be
// a directory (not a file). The directory can have a trailing
// slash or not.
// Example: "/home/seans/project" or "/home/seans/project/"
// NOT "/home/seans/project/file.yaml".
func (l *loaderImpl) New(newRoot string) (Loader, error) {
	if !l.fLoader.IsAbsPath(l.root, newRoot) {
		return nil, fmt.Errorf("Not abs path: l.root='%s', loc='%s'\n", l.root, newRoot)
	}
	root, err := l.fLoader.FullLocation(l.root, newRoot)
	if err != nil {
		return nil, err
	}
	return &loaderImpl{root: root, fLoader: l.fLoader}, nil
}

// Load returns all the bytes read from location or an error.
// "location" can be an absolute path, or if relative, full location is
// calculated from the Root().
func (l *loaderImpl) Load(location string) ([]byte, error) {
	fullLocation, err := l.fLoader.FullLocation(l.root, location)
	if err != nil {
		fmt.Printf("Trouble in fulllocation: %v\n", err)
		return nil, err
	}
	return l.fLoader.Load(fullLocation)
}

// GlobLoad returns a map from path to bytes read from the location or an error.
// "location" can be an absolute path, or if relative, full location is
// calculated from the Root().
func (l *loaderImpl) GlobLoad(location string) (map[string][]byte, error) {
	fullLocation, err := l.fLoader.FullLocation(l.root, location)
	if err != nil {
		return nil, err
	}
	return l.fLoader.GlobLoad(fullLocation)
}
