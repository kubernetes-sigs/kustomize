// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestStringPrefixerPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin(
			"someteam.example.com", "v1", "StringPrefixer")
	defer th.Reset()

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

	th.AssertActualEqualsExpectedNoIdAnnotations(m, `
apiVersion: apps/v1
kind: MeatBall
metadata:
  name: wowsa-meatball
`)
}
