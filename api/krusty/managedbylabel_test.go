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
    app.kubernetes.io/managed-by: kustomize-v444.333.222
  name: myService
spec:
  ports:
  - port: 7002
`

func TestAddManagedbyLabel(t *testing.T) {
	tests := []struct {
		kustFile      string
		managedByFlag bool
		expected      string
	}{
		{
			kustFile: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
`,
			managedByFlag: true,
			expected:      expected,
		},
		{
			kustFile: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
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
