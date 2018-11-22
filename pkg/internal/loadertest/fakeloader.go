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

// Package loadertest holds a fake for the Loader interface.
package loadertest

import (
	"log"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/loader"
)

// FakeLoader encapsulates the delegate Loader and the fake file system.
type FakeLoader struct {
	fs       fs.FileSystem
	delegate ifc.Loader
}

// NewFakeLoader returns a Loader that uses a fake filesystem.
// The argument should be an absolute file path.
func NewFakeLoader(initialDir string) FakeLoader {
	// Create fake filesystem and inject it into initial Loader.
	fSys := fs.MakeFakeFS()
	fSys.Mkdir(initialDir)
	ldr, err := loader.NewLoader(initialDir, fSys)
	if err != nil {
		log.Fatalf("Unable to make loader: %v", err)
	}
	return FakeLoader{fs: fSys, delegate: ldr}
}

// AddFile adds a fake file to the file system.
func (f FakeLoader) AddFile(fullFilePath string, content []byte) error {
	return f.fs.WriteFile(fullFilePath, content)
}

// AddDirectory adds a fake directory to the file system.
func (f FakeLoader) AddDirectory(fullDirPath string) error {
	return f.fs.Mkdir(fullDirPath)
}

// Root returns root.
func (f FakeLoader) Root() string {
	return f.delegate.Root()
}

// New creates a new loader from a new root.
func (f FakeLoader) New(newRoot string) (ifc.Loader, error) {
	l, err := f.delegate.New(newRoot)
	if err != nil {
		return nil, err
	}
	return FakeLoader{fs: f.fs, delegate: l}, nil
}

// Load performs load from a given location.
func (f FakeLoader) Load(location string) ([]byte, error) {
	return f.delegate.Load(location)
}

// Cleanup does nothing
func (f FakeLoader) Cleanup() error {
	return nil
}
