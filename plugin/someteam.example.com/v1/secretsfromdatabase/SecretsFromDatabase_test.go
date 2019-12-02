// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSecretsFromDatabasePlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "SecretsFromDatabase")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: SecretsFromDatabase
metadata:
  name: forbiddenValues
  namespace: production
keys:
- ROCKET
- VEGETABLE
`)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  ROCKET: U2F0dXJuVg==
  VEGETABLE: Y2Fycm90
kind: Secret
metadata:
  name: forbiddenValues
  namespace: production
type: Opaque
`)
}
