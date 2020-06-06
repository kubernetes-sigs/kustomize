// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestInit(t *testing.T) {
	tests := []test{
		{
			name: "init",
			args: []string{"cfg", "init"},
			expectedFiles: map[string]string{
				"Krmfile": `
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
`,
			},
		},

		{
			name: "init",
			args: []string{"cfg", "init", "."},
			expectedFiles: map[string]string{
				"Krmfile": `
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
`,
			},
		},
	}
	runTests(t, tests)
}
