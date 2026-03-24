// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// TestMergeKeys verifies that mergeKeys declared in kustomization.yaml causes
// strategic merge patches to merge list items by key instead of replacing the
// list, even for CRD resources that have no registered OpenAPI schema.
func TestMergeKeys(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	// Base resource with a CRD-like list field.
	th.WriteF("base/resource.yaml", `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: BASE_VAR
    value: base-value
`)
	th.WriteK("base", `
resources:
- resource.yaml
`)

	// Overlay patch that adds a new env var.
	th.WriteF("overlay/patch.yaml", `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: OVERLAY_VAR
    value: overlay-value
`)
	th.WriteK("overlay", `
resources:
- ../base

patches:
- path: patch.yaml

mergeKeys:
- kind: MyApp
  group: example.com
  path: spec/env
  key: name
`)

	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: OVERLAY_VAR
    value: overlay-value
  - name: BASE_VAR
    value: base-value
`)
}
