// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Simplest minimal reproduction of the bug I could make
func writeSimplestBase(th kusttest_test.Harness) {
	th.WriteK("base", `
resources:
- simple.yaml

patches:
- path: nothing-patch.yaml
- path: nothing-patch.yaml
`)
	th.WriteF("base/simple.yaml", `
apiVersion: 1
kind: dnc
metadata:
  name: dnc
thing: null
`)
	th.WriteF("base/nothing-patch.yaml", `
apiVersion: 1
kind: dnc
metadata:
  name: dnc
`)
}

func TestSimplestBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeSimplestBase(th)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: 1
kind: dnc
metadata:
  name: dnc
thing: null
`)
}

// More complex/more interesting reproduction of the bug showing inconsistent manifestation
// Only the null `fakeKey` under `spec.template.spec.containers[name="someName"] is transformed
// to "null".  This case may be useful in tracking down the underlying bug
func writeComplexBase(th kusttest_test.Harness) {
	th.WriteK("base", `
resources:
- complex.yaml

patches:
- path: nothing-patch.yaml
- path: nothing-patch.yaml
`)
	th.WriteF("base/complex.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: fakests
spec:
  template:
    spec:
      containers:
      - fakeKey: null
        name: someName
spec-fake:
  template:
    spec:
      containers:
      - fakeKey: null
        name: someName
`)
	th.WriteF("base/nothing-patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: fakests
`)
}

func TestComplexBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComplexBase(th)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: fakests
spec:
  template:
    spec:
      containers:
      - fakeKey: null
        name: someName
spec-fake:
  template:
    spec:
      containers:
      - fakeKey: null
        name: someName
`)
}
