// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestStringPrefixerPlugin(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "StringPrefixer")
	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app")

	m := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: StringPrefixer
metadata:
  name: wowsa
`,
		`apiVersion: apps/v1
kind: MeatBall
metadata:
  name: meatball
`)

	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: MeatBall
metadata:
  name: wowsa-meatball
`)
}
