// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestBasicIO_1(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	opts := th.MakeDefaultOptions()
	if !opts.UseKyaml {
		// This test won't pass under apimachinery, because in the bowels of
		// that code (see GetAnnotations in v0.17.0 of
		// k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructured.go)
		// an error returned from NestedStringMap is discarded, and an
		// empty annotation map is silently returned, making this test fail
		// The swallowed error arises from code like:
		//   var v interface{}
		//   v = true
		//   if str, ok := v.(string); ok {
		//     save the value in a map (doesn't happen)
		//   } else {
		//     return an error (that is then ignored by GetAnnotations)
		//   }
		// The error happens when any annotation value can be interpreted as
		// a boolean or number.  Such annotations cannot be successfully applied
		// to an object in a cluster unless they are quoted.
		t.SkipNow()
	}
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
	m := th.Run(".", opts)
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
