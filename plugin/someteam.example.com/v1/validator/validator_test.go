// +build notravis

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

func TestValidatorHappy(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin("someteam.example.com", "v1", "Validator")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
`,
		`apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    foo: bar
  name: some-cm
data:
  foo: bar
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    foo: bar
  name: some-cm
`)
}

func TestValidatorUnHappy(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin("someteam.example.com", "v1", "Validator")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	err := th.ErrorFromLoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
`,
		`apiVersion: v1
kind: ConfigMap
metadata:
  annotations: {}
  name: some-cm
data:
- foo: bar
`)
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "failure in plugin configured via") {
		t.Fatalf("unexpected error: %v", err)
	}
}
