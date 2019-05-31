// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/plugin"
)

func TestReplicaCountTransformer(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "ReplicaCountTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ReplicaCountTransformer
metadata:
  name: notImportantHere
replica:
  name: deploy1
  count: 23
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  replicas: 1
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  replicas: 23
`)
}
