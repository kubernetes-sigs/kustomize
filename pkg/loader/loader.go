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
	"path/filepath"

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

// Loader interface exposes methods to read bytes.
type Loader interface {
	// Root returns the root location for this Loader.
	Root() string
	// New returns Loader located at newRoot.
	New(newRoot string) (Loader, error)
	// Load returns the bytes read from the location or an error.
	Load(location string) ([]byte, error)
	// Cleanup cleans the loader
	Cleanup() error
}

// NewLoader returns a Loader given a target
// The target can be a local disk directory or a github Url
func NewLoader(target string, fSys fs.FileSystem) (Loader, error) {
	if isRepoUrl(target) {
		return newGithubLoader(target, fSys)
	}

	l := NewFileLoader(fSys)
	absPath, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}

	if !l.IsAbsPath(l.root, absPath) {
		return nil, fmt.Errorf("Not abs path: l.root='%s', loc='%s'\n", l.root, absPath)
	}
	root, err := l.fullLocation(l.root, absPath)
	if err != nil {
		return nil, err
	}
	return newFileLoaderAtRoot(root, l.fSys), nil
}
