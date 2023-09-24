// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove //nolint:testpackage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRemoveSecret(t *testing.T) {
	const secretName01 = "example-secret-01"
	const secretName02 = "example-secret-02"

	tests := map[string]struct {
		input          string
		args           []string
		expectedOutput string
		expectedErr    string
	}{
		"happy path": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: %s
  files:
  - longsecret.txt
`, secretName01),
			args: []string{secretName01},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"multiple": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: %s
  files:
  - longsecret.txt
- name: %s
  files:
  - longsecret.txt
`, secretName01, secretName02),
			args: []string{
				fmt.Sprintf("%s,%s", secretName01, secretName02),
			},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		"miss": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: %s
  files:
  - longsecret.txt
`, secretName01),
			args:        []string{"foo"},
			expectedErr: "no specified secret(s) were found",
		},
		"no secret name specified": {
			args:        []string{},
			expectedErr: "at least one secret name must be specified",
		},
		"too many secret names specified": {
			args:        []string{"test1", "test2"},
			expectedErr: "too many arguments",
		},
		"one existing and one non-existing": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: %s
  files:
  - application.properties
`, secretName01),
			args: []string{fmt.Sprintf("%s,%s", secretName01, "foo")},
			expectedOutput: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))
			cmd := newCmdRemoveSecret(fSys)
			err := cmd.RunE(cmd, tc.args)

			if tc.expectedErr != "" {
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
