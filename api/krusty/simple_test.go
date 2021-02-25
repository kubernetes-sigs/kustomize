// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSimple1(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("dep.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`)
	th.WriteF("patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 999
`)

	th.WriteK(".", `
resources:
- dep.yaml
patchesStrategicMerge:
- patch.yaml
`)
	m := th.Run(".", th.MakeDefaultOptions())

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 999
`)
}
