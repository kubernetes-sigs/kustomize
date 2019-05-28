// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/plugin"
)

func TestNameTransformer(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "NameTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: NameTransformer
metadata:
  name: notImportantHere
prefix: baked-
suffix: -pie
fieldSpecs:
  - path: metadata/name
`, `
apiVersion: v1
kind: Service
metadata:
  name: apple
spec:
  ports:
  - port: 7002
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: Service
metadata:
  name: baked-apple-pie
spec:
  ports:
  - port: 7002
`)
}
