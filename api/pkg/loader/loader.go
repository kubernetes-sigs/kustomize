// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package pkg has all the helpers to interact with the api.
package loader

import (
	"sigs.k8s.io/kustomize/api/internal/loader"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// NewFileLoaderAtCwd returns a loader that loads from PWD.
// A convenience for kustomize edit commands.
func NewFileLoaderAtCwd(fSys filesys.FileSystem) *loader.FileLoader {
	return loader.NewLoaderOrDie(
		loader.RestrictionRootOnly, fSys, filesys.SelfDir)
}

// NewFileLoaderAtRoot returns a loader that loads from "/".
// A convenience for tests.
func NewFileLoaderAtRoot(fSys filesys.FileSystem) *loader.FileLoader {
	return loader.NewLoaderOrDie(
		loader.RestrictionRootOnly, fSys, filesys.Separator)
}
