// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Glob results are written into kustomization files, which use "/" on
// every platform. A key taken from a file source is the part after the
// last "/", so `dir\fa1` would make "dir\fa1" the key instead of "fa1".
func TestGlobPatterns_Windows(t *testing.T) {
	t.Chdir(t.TempDir())
	fSys := filesys.MakeFsOnDisk()
	require.NoError(t, fSys.MkdirAll("dir"))
	require.NoError(t, fSys.WriteFile(`dir\fa1`, []byte{}))

	for _, pattern := range []string{`dir\fa*`, "dir/fa*"} {
		files, err := GlobPatterns(fSys, []string{pattern})
		require.NoError(t, err)
		assert.Equal(t, []string{"dir/fa1"}, files, pattern)

		files, err = GlobPatternsWithLoader(fSys, nil, []string{pattern}, false)
		require.NoError(t, err)
		assert.Equal(t, []string{"dir/fa1"}, files, pattern)
	}
}
