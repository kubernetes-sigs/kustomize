// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestCalvinDuplicatorPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "CalvinDuplicator")
	defer th.Reset()

	m := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: CalvinDuplicator
metadata:
  name: whatever
count: 3
name: calvin
`,
		`apiVersion: apps/v1
kind: Deployment
metadata:
  name: hobbes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calvin
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: susie
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hobbes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calvin-1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calvin-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calvin-3
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: susie
`)
}
