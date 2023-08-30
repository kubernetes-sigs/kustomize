// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Setup sets up a file system on disk and directory that is cleaned after
// test completion.
func Setup(t *testing.T) (filesys.FileSystem, filesys.ConfirmedDir) {
	t.Helper()

	fSys := filesys.MakeFsOnDisk()
	dir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = fSys.RemoveAll(dir.String())
	})
	return fSys, dir
}

// CreateKustDir creates a file system on disk and a new directory
// that holds a kustomization file with content. The directory is removed on
// test completion.
func CreateKustDir(t *testing.T, content string) (filesys.FileSystem, filesys.ConfirmedDir) {
	t.Helper()

	fSys, tmpDir := Setup(t)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(content)))
	return fSys, tmpDir
}
