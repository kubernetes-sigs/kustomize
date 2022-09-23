// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestKeepEmptyArray(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("resources.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing123
spec:
  replicas: 1
  selector: null
  template:
    spec:
      containers:
      - name: event
        image: testing123
        imagePullPolicy: IfNotPresent
      imagePullSecrets: []`)

	th.WriteK(".", `
resources:
- resources.yaml`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing123
spec:
  replicas: 1
  selector: null
  template:
    spec:
      containers:
      - image: testing123
        imagePullPolicy: IfNotPresent
        name: event
      imagePullSecrets: []
`)
}
