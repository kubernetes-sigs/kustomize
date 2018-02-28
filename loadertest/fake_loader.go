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

package loadertest

import (
	"os"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
	"k8s.io/kubectl/pkg/loader"
)

// FakeLoader encapsulates the delegate Loader and the fake file system.
type FakeLoader struct {
	fs       fs.FileSystem
	delegate loader.Loader
}

// NewFakeLoader returns a Loader that delegates calls, and encapsulates
// a fake file system that the Loader reads from. "initialDir" parameter
// must be an full, absolute directory (trailing slash doesn't matter).
func NewFakeLoader(initialDir string) FakeLoader {
	// Create fake filesystem and inject it into initial Loader.
	fakefs := fs.MakeFakeFS()
	var schemes []loader.SchemeLoader
	schemes = append(schemes, loader.NewFileLoader(fakefs))
	rootLoader := loader.Init(schemes)
	loader, _ := rootLoader.New(initialDir)
	return FakeLoader{fs: fakefs, delegate: loader}
}

// Adds a fake file to the file system.
func (f FakeLoader) AddFile(fullFilePath string, content []byte) error {
	return f.fs.WriteFile(fullFilePath, content)
}

// Adds a fake directory to the file system.
func (f FakeLoader) AddDirectory(fullDirPath string, mode os.FileMode) error {
	return f.fs.Mkdir(fullDirPath, mode)
}

func (f FakeLoader) Root() string {
	return f.delegate.Root()
}

func (f FakeLoader) New(newRoot string) (loader.Loader, error) {
	return f.delegate.New(newRoot)
}

func (f FakeLoader) Load(location string) ([]byte, error) {
	return f.delegate.Load(location)
}
