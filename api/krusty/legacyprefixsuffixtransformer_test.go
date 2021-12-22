// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestLegacyPrefixSuffixTransformer(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
transformers:
- |-
  apiVersion: builtin
  kind: PrefixSuffixTransformer
  metadata:
    name: notImportantHere
  prefix: baked-
  suffix: -pie
  fieldSpecs:
  - path: metadata/name
`)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: apple
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  name: baked-apple-pie
`)
}
