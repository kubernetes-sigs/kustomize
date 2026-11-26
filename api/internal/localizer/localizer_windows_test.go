// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// The localized kustomization must use "/" on every platform, so that it
// builds the same way wherever it is copied to. Walking and cleaning paths
// on Windows yields `a\b`.
func TestLocalizeWritesSlashPaths_Windows(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	for file, content := range map[string]string{
		"kustomization.yaml": `patches:
- path: a/b/patch.yaml
resources:
- c/d
helmGlobals:
  chartHome: charts/home
`,
		"a/b/patch.yaml":               podConfiguration,
		"c/d/kustomization.yaml":       "namePrefix: test-\n",
		"charts/home/name/values.yaml": "",
	} {
		path := filepath.Join(target, filepath.FromSlash(file))
		require.NoError(t, fSys.MkdirAll(filepath.Dir(path)))
		require.NoError(t, fSys.WriteFile(path, []byte(content)))
	}

	dst := filepath.Join(dir, "dst")
	_, err := Run(target, target, dst, fSys)
	require.NoError(t, err)

	content, err := fSys.ReadFile(filepath.Join(dst, "kustomization.yaml"))
	require.NoError(t, err)
	require.Equal(t, `helmGlobals:
  chartHome: charts/home
patches:
- path: a/b/patch.yaml
resources:
- c/d
`, string(content))
}
