// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestDatePrefixerPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "DatePrefixer")
	defer th.Reset()

	m := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: DatePrefixer
metadata:
  name: whatever
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
  name: 2018-05-11-meatball
`)
}
