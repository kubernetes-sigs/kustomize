// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestPodDisruptionBudgetBasics(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("pdbLiteral.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: pdbLiteral
spec:
  maxUnavailable: 90
`)
	th.WriteF("pdbPercentage.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: pdbPercentage
spec:
  maxUnavailable: 90%
`)
	th.WriteK(".", `
resources:
- pdbLiteral.yaml
- pdbPercentage.yaml
`)
	m := th.Run(".", th.MakeDefaultOptions())
	// In a PodDisruptionBudget, the fields maxUnavailable
	// minAvailable are mutually exclusive, and both can hold
	// either an integer, i.e. 10, or string that has to be
	// an int followed by a percent sign, e.g. 10%.
	th.AssertActualEqualsExpected(m, `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: pdbLiteral
spec:
  maxUnavailable: 90
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: pdbPercentage
spec:
  maxUnavailable: 90%
`)
}

func TestPodDisruptionBudgetMerging(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
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
	m := th.Run(".", th.MakeDefaultOptions())
	// In a PodDisruptionBudget, the fields maxUnavailable
	// minAvailable are mutually exclusive, and both can hold
	// either an integer, i.e. 10, or string that has to be
	// an int followed by a percent sign, e.g. 10%.
	// In the former case - bare integer - they should NOT be quoted
	// as the api server will reject it.  In the latter case with
	// the percent sign, quotes can be added and the API server will
	// accept it, but they don't have to be added.
	th.AssertActualEqualsExpected(
		m, `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    faceit-pdb: default
  name: championships-api
spec:
  maxUnavailable: 1
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    faceit-pdb: default
  name: championships-api-2
spec:
  maxUnavailable: 1
`)
}
