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
	"sigs.k8s.io/kustomize/pkg/git"
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
	// Loader that spawned this loader.
	// Used to avoid cycles.
	referrer *fileLoader
	// An absolute, cleaned path to a directory.
	// The Load function reads from this directory,
	// or directories below it.
	root fs.ConfirmedDir
	// URI, if any, used for a download into root.
	// TODO(monopole): use non-string type.
	uri string
	// File system utilities.
	fSys fs.FileSystem
	// Used to clone repositories.
	cloner git.Cloner
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
	return l.root.String()
}

func newLoaderOrDie(fSys fs.FileSystem, path string) *fileLoader {
	l, err := newLoaderAtConfirmedDir(
		path, fSys, nil, git.ClonerUsingGitExec)
	if err != nil {
		log.Fatalf("unable to make loader at '%s'; %v", path, err)
	}
	return l
}

// newLoaderAtConfirmedDir returns a new fileLoader with given root.
func newLoaderAtConfirmedDir(
	possibleRoot string, fSys fs.FileSystem,
	referrer *fileLoader, cloner git.Cloner) (*fileLoader, error) {
	if possibleRoot == "" {
		return nil, fmt.Errorf(
			"loader root cannot be empty")
	}
	root, f, err := fSys.CleanedAbs(possibleRoot)
	if err != nil {
		return nil, fmt.Errorf(
			"absolute path error in '%s' : %v", possibleRoot, err)
	}
	if f != "" {
		return nil, fmt.Errorf(
			"got file '%s', but '%s' must be a directory to be a root",
			f, possibleRoot)
	}
	if referrer != nil {
		if err := referrer.errIfArgEqualOrHigher(root); err != nil {
			return nil, err
		}
	}
	return &fileLoader{
		root:     root,
		referrer: referrer,
		fSys:     fSys,
		cloner:   cloner,
		cleaner:  func() error { return nil },
	}, nil
}

// New returns a new Loader, rooted relative to current loader,
// or rooted in a temp directory holding a git repo clone.
func (l *fileLoader) New(path string) (ifc.Loader, error) {
	if path == "" {
		return nil, fmt.Errorf("new root cannot be empty")
	}
	if git.IsRepoUrl(path) {
		// Avoid cycles.
		if err := l.errIfPreviouslySeenUri(path); err != nil {
			return nil, err
		}
		return newLoaderAtGitClone(path, l.fSys, l.referrer, l.cloner)
	}
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("new root '%s' cannot be absolute", path)
	}
	return newLoaderAtConfirmedDir(
		l.root.Join(path), l.fSys, l, l.cloner)
}

// newLoaderAtGitClone returns a new Loader pinned to a temporary
// directory holding a cloned git repo.
func newLoaderAtGitClone(
	uri string, fSys fs.FileSystem,
	referrer *fileLoader, cloner git.Cloner) (ifc.Loader, error) {
	tmpDirForRepo, pathInRepo, err := cloner(uri)
	if err != nil {
		return nil, err
	}
	root, f, err := fSys.CleanedAbs(
		filepath.Join(tmpDirForRepo, pathInRepo))
	if err != nil {
		return nil, err
	}
	if f != "" {
		return nil, fmt.Errorf(
			"'%s' refers to file '%s'; expecting directory", pathInRepo, f)
	}
	return &fileLoader{
		root:     root,
		referrer: referrer,
		uri:      uri,
		fSys:     fSys,
		cloner:   cloner,
		cleaner:  func() error { return fSys.RemoveAll(tmpDirForRepo) },
	}, nil
}

// errIfArgEqualOrHigher tests whether the argument,
// is equal to or above the root of any ancestor.
func (l *fileLoader) errIfArgEqualOrHigher(
	candidateRoot fs.ConfirmedDir) error {
	if l.root.HasPrefix(candidateRoot) {
		return fmt.Errorf(
			"cycle detected: candidate root '%s' contains visited root '%s'",
			candidateRoot, l.root)
	}
	if l.referrer == nil {
		return nil
	}
	return l.referrer.errIfArgEqualOrHigher(candidateRoot)
}

// TODO(monopole): Distinguish branches?
// I.e. Allow a distinction between git URI with
// path foo and tag bar and a git URI with the same
// path but a different tag?
func (l *fileLoader) errIfPreviouslySeenUri(uri string) error {
	if strings.HasPrefix(l.uri, uri) {
		return fmt.Errorf(
			"cycle detected: URI '%s' referenced by previous URI '%s'",
			uri, l.uri)
	}
	if l.referrer == nil {
		return nil
	}
	return l.referrer.errIfPreviouslySeenUri(uri)
}

// Load returns content of file at the given relative path,
// else an error.  The path must refer to a file in or
// below the current root.
func (l *fileLoader) Load(path string) ([]byte, error) {
	if filepath.IsAbs(path) {
		return nil, l.loadOutOfBounds(path)
	}
	d, f, err := l.fSys.CleanedAbs(l.root.Join(path))
	if err != nil {
		return nil, err
	}
	if f == "" {
		return nil, fmt.Errorf(
			"'%s' must be a file (got d='%s')", path, d)
	}
	if !d.HasPrefix(l.root) {
		return nil, l.loadOutOfBounds(path)
	}
	return l.fSys.ReadFile(d.Join(f))
}

func (l *fileLoader) loadOutOfBounds(path string) error {
	return fmt.Errorf(
		"security; file '%s' is not in or below '%s'",
		path, l.root)
}

// Cleanup runs the cleaner.
func (l *fileLoader) Cleanup() error {
	return l.cleaner()
}
