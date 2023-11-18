// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"testing"

	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRemoveSecret(t *testing.T) {
	tests := map[string]struct {
		input          string
		args           []string
		expectedOutput string
		wantErr        bool
		expectedErr    string
	}{
		"removes a secret successfully": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  files:
  - longsecret.txt
`,
			args: []string{"test-secret-1"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"removes multiple secrets successfully": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  files:
  - longsecret.txt
- name: test-secret-2
  files:
  - longsecret.txt
`,
			args: []string{"test-secret-1,test-secret-2"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"removes a secret successfully when single secret name is specified, it exists, but is not the only secret present": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  namespace: test-ns
  files:
  - longsecret.txt
- name: test-secret-2
  namespace: default
  literals:
  - test-key=test-secret
`,
			args:    []string{"test-secret-1", "--namespace=test-ns"},
			wantErr: false,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - test-key=test-secret
  name: test-secret-2
  namespace: default
`,
		},
		"succeeds when one secret name exists and one doesn't exist": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  files:
  - application.properties
`,
			args: []string{"test-secret-1,test-secret-2"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"succeeds when one secret name exists in the specified namespace": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret
  namespace: test-ns
  files:
  - application.properties
`,
			args: []string{"test-secret", "--namespace=test-ns"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"handles empty namespace as default in the args": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret
  namespace: default
  files:
  - application.properties
`,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
			wantErr: false,
			args:    []string{"test-secret"},
		},
		"handles empty namespace as default in the kustomization file": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret
  files:
  - application.properties
`,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
			wantErr: false,
			args:    []string{"test-secret", "--namespace=default"},
		},
		"fails when single secret name is specified and doesn't exist": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  files:
  - longsecret.txt
`,
			args:        []string{"foo"},
			wantErr:     true,
			expectedErr: "no specified secret(s) were found",
		},
		"fails when single secret name is specified and doesn't exist in the specified namespace": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  namespace: test-ns
  files:
  - longsecret.txt
`,
			args:        []string{"test-secret-1"},
			wantErr:     true,
			expectedErr: "no specified secret(s) were found",
		},

		"fails when single secret name is specified and doesn't exist in the specified namespace, and neither namespace is the default one": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: test-secret-1
  namespace: test-ns
  files:
  - longsecret.txt
`,
			args:        []string{"test-secret-1", "--namespace=other-ns"},
			wantErr:     true,
			expectedErr: "no specified secret(s) were found",
		},
		"fails when no secret name is specified": {
			args:        []string{},
			wantErr:     true,
			expectedErr: "at least one secret name must be specified",
		},
		"fails when too many secret names are specified": {
			args:        []string{"test1", "test2"},
			wantErr:     true,
			expectedErr: "too many arguments",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))
			cmd := newCmdRemoveSecret(fSys)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			require.NoError(t, err)
			content, err := testutils_test.ReadTestKustomization(fSys)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOutput, string(content))
		})
	}
}
