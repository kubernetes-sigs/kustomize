// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestBasicIO_1(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
`)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    port: 8080
    happy: true
    color: green
  name: demo
spec:
  clusterIP: None
`)
	m := th.Run(".", th.MakeDefaultOptions())
	// The annotations are sorted by key, hence the order change.
	// Quotes are added intentionally.
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    color: green
    happy: "true"
    port: "8080"
  name: demo
spec:
  clusterIP: None
`)
}

func TestBasicIO_2(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	opts := th.MakeDefaultOptions()
	th.WriteK(".", `
resources:
- service.yaml
`)
	// All the annotation values are quoted in the input.
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    port: "8080"
    happy: "true"
    color: green
  name: demo
spec:
  clusterIP: None
`)
	m := th.Run(".", opts)
	// The annotations are sorted by key, hence the order change.
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    color: green
    happy: "true"
    port: "8080"
  name: demo
spec:
  clusterIP: None
`)
}
