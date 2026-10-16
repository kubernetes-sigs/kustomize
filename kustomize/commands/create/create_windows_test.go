// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Detected resources are written into the kustomization file, which uses
// "/" on every platform, while walking a directory yields `sub\test.yaml`.
func TestDetectResources_Windows(t *testing.T) {
	t.Chdir(t.TempDir())
	fSys := filesys.MakeFsOnDisk()
	require.NoError(t, fSys.MkdirAll("sub"))
	require.NoError(t, fSys.WriteFile(`sub\test.yaml`, []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test`)))

	paths, err := detectResources(fSys, factory, ".", true)
	require.NoError(t, err)
	assert.Equal(t, []string{"sub/test.yaml"}, paths)
}
