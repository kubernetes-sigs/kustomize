// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"testing"

	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRemoveConfigMap(t *testing.T) {
	tests := map[string]struct {
		input          string
		args           []string
		expectedOutput string
		wantErr        bool
		expectedErr    string
	}{
		"removes a configmap successfully": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  files:
  - application.properties
`,
			args: []string{"test-cm-1"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"removes multiple configmaps successfully": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  files:
  - application.properties
- name: test-cm-2
  files:
  - application.properties
`,
			args: []string{"test-cm-1,test-cm-2"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"removes a configmap successfully when single configmap name is specified, it exists, but is not the only configmap present": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  namespace: test-ns
  files:
  - application.properties
- name: test-cm-2
  namespace: default
  literals:
  - test-key=test-value
`,
			args:    []string{"test-cm-1", "--namespace=test-ns"},
			wantErr: false,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test-key=test-value
  name: test-cm-2
  namespace: default
`,
		},
		"succeeds when one configmap name exists and one doesn't exist": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  files:
  - application.properties
`,
			args: []string{"test-cm-1,foo"},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"succeeds when one configmap name exists in the specified namespace": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm
  namespace: test-ns
  files:
  - application.properties
`,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
			wantErr: false,
			args:    []string{"test-cm", "--namespace=test-ns"},
		},
		"handles empty namespace as default in the args": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm
  namespace: default
  files:
  - application.properties
`,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
			wantErr: false,
			args:    []string{"test-cm"},
		},
		"handles empty namespace as default in the kustomization file": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm
  files:
  - application.properties
`,
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
			wantErr: false,
			args:    []string{"test-cm", "--namespace=default"},
		},
		"fails when single configmap name is specified and doesn't exist": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  files:
  - application.properties
`,
			args:        []string{"foo"},
			wantErr:     true,
			expectedErr: "no specified configmap(s) were found",
		},
		"fails when single configmap name is specified and doesn't exist in the specified namespace": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  namespace: test-ns
  files:
  - application.properties
`,
			args:        []string{"test-cm-1"},
			wantErr:     true,
			expectedErr: "no specified configmap(s) were found",
		},

		"fails when single configmap name is specified and doesn't exist in the specified namespace, and neither namespace is the default one": {
			input: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: test-cm-1
  namespace: test-ns
  files:
  - application.properties
`,
			args:        []string{"test-cm-1", "--namespace=other-ns"},
			wantErr:     true,
			expectedErr: "no specified configmap(s) were found",
		},
		"fails when no configmap name is specified": {
			args:        []string{},
			wantErr:     true,
			expectedErr: "at least one configmap name must be specified",
		},
		"fails when too many configmap names are specified": {
			args:        []string{"test1", "test2"},
			wantErr:     true,
			expectedErr: "too many arguments",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))
			cmd := newCmdRemoveConfigMap(fSys)
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
