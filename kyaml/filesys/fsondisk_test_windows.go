// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func getOSRoot(t *testing.T) string {
	t.Helper()

	sysDir, err := windows.GetSystemDirectory()
	require.NoError(t, err)
	return filepath.VolumeName(sysDir) + `\`
}
