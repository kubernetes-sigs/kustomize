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
	"path/filepath"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

// Implements internal interface schemeLoader.
type fileLoader struct {
	fs fs.FileSystem
}

func newFileLoader(fs fs.FileSystem) (schemeLoader, error) {
	return &fileLoader{fs: fs}, nil
}

// Join the root path with the location path.
func (l *fileLoader) fullLocation(root string, location string) string {
	fullLocation := location
	if !filepath.IsAbs(location) {
		fullLocation = filepath.Join(root, location)
	}
	return fullLocation
}

// Load returns the bytes from reading a file at fullFilePath.
// Implements the Loader interface.
func (l *fileLoader) load(fullFilePath string) ([]byte, error) {
	// TODO: Check that fullFilePath is an absolute file path.
	return l.fs.ReadFile(fullFilePath)
}
