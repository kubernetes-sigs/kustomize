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
	"log"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

// fileLoader loads files from a file system.
// It has a notion of a current working directory, called 'root',
// that is independent from the current working directory of the
// process. When it loads a file from a relative path, the load
// is done relative to this root, not the process CWD.
type fileLoader struct {
	// Previously visited directories, tracked to avoid cycles.
	// The last entry is the current root.
	roots []string
	// File system utilities.
	fSys fs.FileSystem
}

// NewFileLoaderAtCwd returns a loader that loads from ".".
func NewFileLoaderAtCwd(fSys fs.FileSystem) *fileLoader {
	return newLoaderOrDie(fSys, ".")
}

// NewFileLoaderAtRoot returns a loader that loads from "/".
func NewFileLoaderAtRoot(fSys fs.FileSystem) *fileLoader {
	return newLoaderOrDie(fSys, "/")
}

// Root returns the absolute path that is prepended to any relative paths
// used in Load.
func (l *fileLoader) Root() string {
	return l.roots[len(l.roots)-1]
}

func newLoaderOrDie(fSys fs.FileSystem, path string) *fileLoader {
	l, err := newFileLoaderAt(fSys, path)
	if err != nil {
		log.Fatalf("unable to make loader at '%s'; %v", path, err)
	}
	return l
}

// newFileLoaderAt returns a new fileLoader with given root.
func newFileLoaderAt(fSys fs.FileSystem, root string) (*fileLoader, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf(
			"no absolute path for '%s' : %v", root, err)
	}
	if !fSys.IsDir(root) {
		return nil, fmt.Errorf("absolute root '%s' must exist", root)
	}
	return &fileLoader{roots: []string{root}, fSys: fSys}, nil
}

// Returns a new Loader, which might be rooted relative to current loader.
func (l *fileLoader) New(root string) (ifc.Loader, error) {
	if root == "" {
		return nil, fmt.Errorf("new root cannot be empty")
	}
	if isRepoUrl(root) {
		return newGithubLoader(root, l.fSys)
	}
	if filepath.IsAbs(root) {
		return l.childLoaderAt(filepath.Clean(root))
	}
	// Get absolute path to squeeze out "..", ".", etc. to check for cycles.
	absRoot, err := filepath.Abs(filepath.Join(l.Root(), root))
	if err != nil {
		return nil, fmt.Errorf(
			"problem joining '%s' and '%s': %v", l.Root(), root, err)
	}
	return l.childLoaderAt(absRoot)
}

// childLoaderAt returns a new fileLoader with given root.
func (l *fileLoader) childLoaderAt(root string) (*fileLoader, error) {
	if !l.fSys.IsDir(root) {
		return nil, fmt.Errorf("absolute root '%s' must exist", root)
	}
	if err := l.seenBefore(root); err != nil {
		return nil, err
	}
	return &fileLoader{roots: append(l.roots, root), fSys: l.fSys}, nil
}

// seenBefore tests whether the current or any previously
// visited root begins with the given path.
// This disallows an overlay from depending on a base positioned
// above it.  There's no good reason to allow this, and to disallow
// it avoid cycles, especially if some future change re-introduces
// globbing to resource and base specification.
func (l *fileLoader) seenBefore(path string) error {
	for _, r := range l.roots {
		if strings.HasPrefix(r, path) {
			return fmt.Errorf(
				"cycle detected: new root '%s' contains previous root '%s'",
				path, r)
		}
	}
	return nil
}

// Load returns content of file at the given relative path.
func (l *fileLoader) Load(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.Root(), path)
	}
	return l.fSys.ReadFile(path)
}

// Cleanup does nothing
func (l *fileLoader) Cleanup() error {
	return nil
}
