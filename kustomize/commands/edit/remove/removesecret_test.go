// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRemoveSecret(t *testing.T) {
	const secretName01 = "example-secret-01"
	const secretName02 = "example-secret-02"

	tests := map[string]struct {
		input       string
		args        []string
		expectedErr string
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
			args: []string{"foo"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))
			cmd := newCmdRemoveSecret(fSys)
			err := cmd.RunE(cmd, tc.args)
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
				content, err := testutils_test.ReadTestKustomization(fSys)
				assert.NoError(t, err)
				for _, opt := range strings.Split(tc.args[0], ",") {
					assert.NotContains(t, string(content), opt)
				}
			}
		})
	}
}
