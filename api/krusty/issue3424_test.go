// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Demonstrate unwanted quotes added to an integer by a strategic merge patch.
func TestIssue3424Basics(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	opts := th.MakeDefaultOptions()
	th.WriteF("pdb-patch.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: generic-pdb
spec:
  maxUnavailable: 1
`)
	th.WriteF("my_file.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: championships-api
  labels:
    faceit-pdb: default
spec:
  maxUnavailable: 100%
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: championships-api-2
  labels:
    faceit-pdb: default
spec:
  maxUnavailable: 100%
`)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: pdb-patch.yaml
  target:
    kind: PodDisruptionBudget
    labelSelector: faceit-pdb=default

resources:
- my_file.yaml
`)
	m := th.Run(".", opts)
	expFmt := `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    faceit-pdb: default
  name: championships-api
spec:
  maxUnavailable: %s
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    faceit-pdb: default
  name: championships-api-2
spec:
  maxUnavailable: %s
`
	th.AssertActualEqualsExpected(m, opts.IfApiMachineryElseKyaml(
		fmt.Sprintf(expFmt, `"1"`, `"1"`),
		fmt.Sprintf(expFmt, `1`, `1`)))
}
