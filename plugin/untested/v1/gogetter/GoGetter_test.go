// +build notravis

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Disabled on travis, because don't want to install go-getter on travis.

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This test requires having the go-getter binary on the PATH.
func TestGoGetter(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "GoGetter")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: GoGetter
metadata:
  name: example
url: github.com/kustless/kustomize-examples.git?ref=adef0a8
`)

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  altGreeting: Good Morning!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: remote-cm
`)
}

func TestGoGetterUrl(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "GoGetter")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: GoGetter
metadata:
  name: example
url: https://github.com/kustless/kustomize-examples/archive/master.zip
subPath: kustomize-examples-master
`)

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  altGreeting: Good Morning!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: remote-cm
`)
}

func TestGoGetterCommand(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "GoGetter")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: GoGetter
metadata:
  name: example
url: github.com/kustless/kustomize-examples.git?ref=adef0a8
command: cat resources.yaml
`)

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  altGreeting: Good Morning!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: cm
`)
}

func TestGoGetterSubPath(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "GoGetter")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: GoGetter
metadata:
  name: example
url: github.com/kustless/kustomize-examples.git?ref=9ca07d2
subPath: dev
command: kustomize build --enable_alpha_plugins
`)

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  altGreeting: Good Morning!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: dev-remote-cm
`)
}
