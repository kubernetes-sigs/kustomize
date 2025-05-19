// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestStableListOrderAfterStrategicMergePatch(t *testing.T) {

	th := kusttest_test.MakeHarness(t)
	th.WriteF("base.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: example
spec:
  containers:
  - name: example
    env:
    - name: ALFA
      value: alfa
    - name: BRAVO
      value: bravo
    - name: CHARLIE
      value: charlie
    - name: DELTA
      value: delta
    - name: LAST
      value: "$(DELTA)$(ALFA)$(DELTA)"
`)

        th.WriteF("patch.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: example
spec:
  containers:
  - name: example
    env:
    - name: LAST
      value: "$(ALFA)$(DELTA)$(ALFA)"
`)
        th.WriteK(".", `
resources:
  - base.yaml
patches:
  - path: patch.yaml
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Pod
metadata:
  name: example
spec:
  containers:
  - name: example
    env:
    - name: ALFA
      value: alfa
    - name: BRAVO
      value: bravo
    - name: CHARLIE
      value: charlie
    - name: DELTA
      value: delta
    - name: LAST
      value: "$(ALFA)$(DELTA)$(ALFA)"
`)
}
