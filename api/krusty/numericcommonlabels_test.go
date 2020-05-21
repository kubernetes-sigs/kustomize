// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This test shows numeric-looking strings correctly handled as strings
func TestNumericCommonLabels(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	// A basic deployment just used to put labels into
	th.WriteF("/app/default/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
`)

	// Combine these custom transformers in one kustomization file.
	// This kustomization file has a string-valued commonLabel that
	// should always be quoted to remain a string
	th.WriteK("/app/default", `
commonLabels:
  version: "1"
resources:
- deployment.yaml
`)

	m := th.Run("/app/default", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    version: "1"
  name: the-deployment
spec:
  selector:
    matchLabels:
      version: "1"
  template:
    metadata:
      labels:
        version: "1"
`)
}
