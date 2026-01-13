// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamespaceShortcutBeforeGenerators_Issue6027(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()

	root := th.GetRoot()
	fs := th.GetFSys()

	dirs := []string{
		filepath.Join(root, "cm-generator", "resources"),
		filepath.Join(root, "base", "configuration"),
		filepath.Join(root, "overlay", "configuration"),
	}
	for _, dir := range dirs {
		require.NoError(t, fs.MkdirAll(dir))
	}

	th.WriteF(filepath.Join(root, "base", "configuration", "general"), "BASE_LAYER_ENV=base\n")
	th.WriteF(filepath.Join(root, "overlay", "configuration", "general"), "OVERLAY_ENV=overlay\n")

	th.WriteK(filepath.Join(root, "cm-generator"), `
resources:
- resources/general.yaml
`)

	th.WriteF(filepath.Join(root, "cm-generator", "resources", "general.yaml"), `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: general-environment
behavior: merge
envs:
- configuration/general
`)

	th.WriteK(filepath.Join(root, "base"), `
generators:
- ../cm-generator
configMapGenerator:
- name: general-environment
  behavior: create
`)

	th.WriteK(filepath.Join(root, "overlay"), `
namespace: abc
generators:
- ../cm-generator
resources:
- ../base
`)

	opts := th.MakeDefaultOptions()
	rm := th.Run(filepath.Join(root, "overlay"), opts)

	hashPattern := regexp.MustCompile(`general-environment-[a-z0-9]+`)
	normalize := func(in []byte) []byte {
		return hashPattern.ReplaceAll(in, []byte("general-environment-HASH"))
	}

	th.AssertActualEqualsExpectedWithTweak(rm, normalize, `
apiVersion: v1
data:
  BASE_LAYER_ENV: base
  OVERLAY_ENV: overlay
kind: ConfigMap
metadata:
  name: general-environment-HASH
  namespace: abc
`)
}
