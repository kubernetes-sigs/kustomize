// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestIssue4388(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- resources.yaml
`)
	th.WriteF("resources.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: testing
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: testing-one
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: testing-two
data:
  key: value
`)
	opts := th.MakeDefaultOptions()
	opts.Reorder = krusty.ReorderOptionLegacy
	m := th.Run(".", opts)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  name: testing
---
apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  name: testing-one
---
apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  name: testing-two
`)
}
