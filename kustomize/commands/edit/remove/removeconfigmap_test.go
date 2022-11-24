// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove //nolint:testpackage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRemoveConfigMap(t *testing.T) {
	const configMapName01 = "example-configmap-01"
	const configMapName02 = "example-configmap-02"

	tests := map[string]struct {
		input       string
		args        []string
		expectedErr string
	}{
		"happy path": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: %s
  files:
  - application.properties
`, configMapName01),
			args: []string{configMapName01},
		},
		"multiple": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: %s
  files:
  - application.properties
- name: %s
  files:
  - application.properties
`, configMapName01, configMapName02),
			args: []string{
				fmt.Sprintf("%s,%s", configMapName01, configMapName02),
			},
		},
		"miss": {
			input: fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: %s
  files:
  - application.properties
`, configMapName01),
			args: []string{"foo"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))
			cmd := newCmdRemoveConfigMap(fSys)
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
