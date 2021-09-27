// +build notravis

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestValidatorHappy(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "Validator")
	defer th.Reset()

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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "Validator")
	defer th.Reset()

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
