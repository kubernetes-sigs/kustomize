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

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

const currentDir = "."

// FileLoader loads files from a file system.
type FileLoader struct {
	fs fs.FileSystem
}

// NewFileLoader returns a new FileLoader.
func NewFileLoader(fs fs.FileSystem) *FileLoader {
	return &FileLoader{fs: fs}
}

// IsAbsPath return true if the location calculated with the root
// and location params a full file path.
func (l *FileLoader) IsAbsPath(root string, location string) bool {
	fullFilePath, err := l.FullLocation(root, location)
	if err != nil {
		return false
	}
	return filepath.IsAbs(fullFilePath)
}

// FullLocation returns some notion of a full path.
// If location is a full file path, then ignore root. If location is relative, then
// join the root path with the location path. Either root or location can be empty,
// but not both. Special case for ".": Expands to current working directory.
// Example: "/home/seans/project", "subdir/bar" -> "/home/seans/project/subdir/bar".
func (l *FileLoader) FullLocation(root string, location string) (string, error) {
	// First, validate the parameters
	if len(root) == 0 && len(location) == 0 {
		return "", fmt.Errorf("unable to calculate full location: root and location empty")
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
func (l *FileLoader) Load(p string) ([]byte, error) {
	return l.fs.ReadFile(p)
}

// GlobLoad returns the map from path to bytes from reading a glob path.
// Implements the Loader interface.
func (l *FileLoader) GlobLoad(p string) (map[string][]byte, error) {
	return l.fs.ReadFiles(p)
}
