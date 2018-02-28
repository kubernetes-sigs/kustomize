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
	"os"
	"path/filepath"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

const currentDir = "."

// Internal implementation of SchemeLoader interface.
type fileLoader struct {
	fs fs.FileSystem
}

// NewFileLoader returns a SchemeLoader to handle a file system.
func NewFileLoader(fs fs.FileSystem) SchemeLoader {
	return &fileLoader{fs: fs}
}

// Is the location calculated with the root and location params a full file path.
func (l *fileLoader) IsScheme(root string, location string) bool {
	fullFilePath, err := l.FullLocation(root, location)
	if err != nil {
		return false
	}
	return filepath.IsAbs(fullFilePath)
}

// If location is a full file path, then ignore root. If location is relative, then
// join the root path with the location path. Either root or location can be empty,
// but not both. Special case for ".": Expands to current working directory.
// Example: "/home/seans/project", "subdir/bar" -> "/home/seans/project/subdir/bar".
func (l *fileLoader) FullLocation(root string, location string) (string, error) {
	// First, validate the parameters
	if len(root) == 0 && len(location) == 0 {
		return "", fmt.Errorf("Unable to calculate full location: root and location empty")
	}
	// Special case current directory, expanding to full file path.
	if location == currentDir {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		location = currentDir
	}
	// Assume the location is a full file path. If not, then join root with location.
	fullLocation := location
	if !filepath.IsAbs(location) {
		fullLocation = filepath.Join(root, location)
	}
	return fullLocation, nil
}

// Load returns the bytes from reading a file at fullFilePath.
// Implements the Loader interface.
func (l *fileLoader) Load(fullFilePath string) ([]byte, error) {
	// Validate path to load from is a full file path.
	if !filepath.IsAbs(fullFilePath) {
		return nil, fmt.Errorf("Attempting to load file without full file path: %s\n", fullFilePath)
	}
	return l.fs.ReadFile(fullFilePath)
}
