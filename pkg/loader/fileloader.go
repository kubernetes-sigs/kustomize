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

// fileLoader is a kustomization's interface to files.
//
// The directory in which a kustomization file sits
// is referred to below as the kustomization's root.
//
// An instance of fileLoader has an immutable root,
// and offers a `New` method returning a new loader
// with a new root.
//
// A kustomization file refers to two kinds of files:
//
// * supplemental data paths
//
//   `Load` is used to visit these paths.
//
//   They must terminate in or below the root.
//
//   They hold things like resources, patches,
//   data for ConfigMaps, etc.
//
// * bases; other kustomizations
//
//   `New` is used to load bases.
//
//   A base can be either a remote git repo URL, or
//   a directory specified relative to the current
//   root. In the former case, the repo is locally
//   cloned, and the new loader is rooted on a path
//   in that clone.
//
//   As loaders create new loaders, a root history
//   is established, and used to disallow:
//
//   - A base that is a repository that, in turn,
//     specifies a base repository seen previously
//     in the loading stack (a cycle).
//
//   - An overlay depending on a base positioned at
//     or above it.  I.e. '../foo' is OK, but '.',
//     '..', '../..', etc. are disallowed.  Allowing
//     such a base has no advantages and encourages
//     cycles, particularly if some future change
//     were to introduce globbing to file
//     specifications in the kustomization file.
//
// These restrictions assure that kustomizations
// are self-contained and relocatable, and impose
// some safety when relying on remote kustomizations,
// e.g. a ConfigMap generator specified to read
// from /etc/passwd will fail.
//
type fileLoader struct {
	// Previously visited absolute directory paths.
	// Tracked to avoid cycles.
	// The last entry is the current root.
	roots []string
	// File system utilities.
	fSys fs.FileSystem
	// Used to clone repositories.
	cloner gitCloner
	// Used to clean up, as needed.
	cleaner func() error
}

// NewFileLoaderAtCwd returns a loader that loads from ".".
func NewFileLoaderAtCwd(fSys fs.FileSystem) *fileLoader {
	return newLoaderOrDie(fSys, ".")
}

// NewFileLoaderAtRoot returns a loader that loads from "/".
func NewFileLoaderAtRoot(fSys fs.FileSystem) *fileLoader {
	return newLoaderOrDie(fSys, string(filepath.Separator))
}

// Root returns the absolute path that is prepended to any
// relative paths used in Load.
func (l *fileLoader) Root() string {
	return l.roots[len(l.roots)-1]
}

func newLoaderOrDie(fSys fs.FileSystem, path string) *fileLoader {
	l, err := newFileLoaderAt(
		path, fSys, []string{}, simpleGitCloner)
	if err != nil {
		log.Fatalf("unable to make loader at '%s'; %v", path, err)
	}
	return l
}

// newFileLoaderAt returns a new fileLoader with given root.
func newFileLoaderAt(
	root string, fSys fs.FileSystem,
	roots []string, cloner gitCloner) (*fileLoader, error) {
	if root == "" {
		return nil, fmt.Errorf(
			"loader root cannot be empty")
	}
	absRoot, f, err := fSys.CleanedAbs(root)
	if err != nil {
		return nil, fmt.Errorf(
			"absolute path error in '%s' : %v", root, err)
	}
	if f != "" {
		return nil, fmt.Errorf(
			"got file '%s', but '%s' must be a directory to be a root",
			f, root)
	}
	if err := isPathEqualToOrAbove(absRoot, roots); err != nil {
		return nil, err
	}
	if !fSys.IsDir(absRoot) {
		return nil, fmt.Errorf(
			"'%s' does not exist or is not a directory", absRoot)
	}
	return &fileLoader{
		roots:   append(roots, absRoot),
		fSys:    fSys,
		cloner:  cloner,
		cleaner: func() error { return nil },
	}, nil
}

// New returns a new Loader, rooted relative to current loader,
// or rooted in a temp directory holding a git repo clone.
func (l *fileLoader) New(path string) (ifc.Loader, error) {
	if path == "" {
		return nil, fmt.Errorf("new root cannot be empty")
	}
	if isRepoUrl(path) {
		// This works well enough for purpose at hand to detect
		// previously visited URLs and thus avoid cycles.
		if err := isPathEqualToOrAbove(path, l.roots); err != nil {
			return nil, err
		}
		return newGitLoader(path, l.fSys, l.roots, l.cloner)
	}
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("new root '%s' cannot be absolute", path)
	}
	return newFileLoaderAt(
		filepath.Join(l.Root(), path), l.fSys, l.roots, l.cloner)
}

// newGitLoader returns a new Loader pinned to a temporary
// directory holding a cloned git repo.
func newGitLoader(
	root string, fSys fs.FileSystem,
	roots []string, cloner gitCloner) (ifc.Loader, error) {
	tmpDirForRepo, pathInRepo, err := cloner(root)
	if err != nil {
		return nil, err
	}
	trueRoot := filepath.Join(tmpDirForRepo, pathInRepo)
	if !fSys.IsDir(trueRoot) {
		return nil, fmt.Errorf(
			"something wrong cloning '%s'; unable to find '%s'",
			root, trueRoot)
	}
	return &fileLoader{
		roots:   append(roots, root, trueRoot),
		fSys:    fSys,
		cloner:  cloner,
		cleaner: func() error { return fSys.RemoveAll(tmpDirForRepo) },
	}, nil
}

// isPathEqualToOrAbove tests whether the 1st argument,
// viewed as a path to a directory, is equal to or above
// any of the paths in the 2nd argument.  It's assumed
// that all paths are cleaned, delinkified and absolute.
func isPathEqualToOrAbove(path string, roots []string) error {
	terminated := path + string(filepath.Separator)
	for _, r := range roots {
		if r == path || strings.HasPrefix(r, terminated) {
			return fmt.Errorf(
				"cycle detected: new root '%s' contains previous root '%s'",
				path, r)
		}
	}
	return nil
}

// Load returns content of file at the given relative path,
// else an error.  The path must refer to a file in or
// below the current Root().
func (l *fileLoader) Load(path string) ([]byte, error) {
	if filepath.IsAbs(path) {
		return nil, l.loadOutOfBounds(path)
	}
	d, f, err := l.fSys.CleanedAbs(
		filepath.Join(l.Root(), path))
	if err != nil {
		return nil, err
	}
	if f == "" {
		return nil, fmt.Errorf(
			"'%s' must be a file (got d='%s')", path, d)
	}
	path = filepath.Join(d, f)
	if !l.isInOrBelowRoot(path) {
		return nil, l.loadOutOfBounds(path)
	}
	return l.fSys.ReadFile(path)
}

// isInOrBelowRoot is true if the argument is in or
// below Root() from purely a path perspective (no
// check for actual file existence). For this to work,
// both the given argument "path" and l.Root() must
// be cleaned, absolute paths, and l.Root() must be
// a directory.  Both conditions enforced elsewhere.
//
// This is tested on linux, but will have trouble
// on other operating systems.  As soon as related
// work is completed in the core filepath package,
// this code should be refactored to use it.
// See:
//   https://github.com/golang/go/issues/18355
//   https://github.com/golang/dep/issues/296
//   https://github.com/golang/dep/blob/master/internal/fs/fs.go#L33
//   https://codereview.appspot.com/5712045
func (l *fileLoader) isInOrBelowRoot(path string) bool {
	if l.Root() == string(filepath.Separator) {
		return true
	}
	return strings.HasPrefix(
		path, l.Root()+string(filepath.Separator))
}

func (l *fileLoader) loadOutOfBounds(path string) error {
	return fmt.Errorf(
		"security; file '%s' is not in or below '%s'",
		path, l.Root())
}

// Cleanup runs the cleaner.
func (l *fileLoader) Cleanup() error {
	return l.cleaner()
}
