// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
)

func TestSomeServiceGeneratorPlugin(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "SomeServiceGenerator")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: SomeServiceGenerator
metadata:
  name: myGenerator
name: my-service
port: "12345"
`)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dev
  name: my-service
spec:
  ports:
  - port: 12345
  selector:
    app: dev
`)
}
