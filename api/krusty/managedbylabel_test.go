// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const expected = `apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/managed-by: kustomize-(test)
  name: myService
spec:
  ports:
  - port: 7002
`

// This test may fail when running on package tests using the go command because `(test)` is set on makefile.
func TestAddManagedbyLabel(t *testing.T) {
	tests := []struct {
		kustFile      string
		managedByFlag bool
		expected      string
	}{
		{
			kustFile: `
resources:
- service.yaml
`,
			managedByFlag: true,
			expected:      expected,
		},
		{
			kustFile: `
resources:
- service.yaml
buildMetadata: [managedByLabel]
`,
			managedByFlag: false,
			expected:      expected,
		},
	}
	for _, tc := range tests {
		th := kusttest_test.MakeHarness(t)
		th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`)
		th.WriteK(".", tc.kustFile)
		options := th.MakeDefaultOptions()
		options.AddManagedbyLabel = tc.managedByFlag
		m := th.Run(".", options)
		th.AssertActualEqualsExpected(m, tc.expected)
	}
}
