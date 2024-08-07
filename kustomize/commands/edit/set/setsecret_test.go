// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The duplicate linter is reporting that setconfigmap_test.go and this file are duplicates, which is not true.
// Disabling lint for these two files specifically to work around that.
//
//nolint:dupl
package set

import (
	"testing"

	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/configmapsecret"
)

func TestFailureCasesEditSetSecret(t *testing.T) {
	testCases := []testutils_test.FailureCase{
		{
			Name: "fails to set a value because key doesn't exist",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - key1=val1
  name: test-secret
  type: Opaque
- literals:
  - key3=val1
  name: test-secret-2
  namespace: test-ns
  type: Opaque
`,
			Args:             []string{"test-secret", "--from-literal=key3=val2"},
			ExpectedErrorMsg: "key 'key3' not found in resource",
		},
		{
			Name: "fails to set a value because secret doesn't exist",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - key1=val1
  name: test-secret
  type: Opaque
- literals:
  - key2=value
  name: another-secret
  namespace: a-namespace
  type: Opaque
`,
			Args:             []string{"test-secret2", "--from-literal=key3=val2"},
			ExpectedErrorMsg: "unable to find Secret with name \"test-secret2\"",
		},
		{
			Name: "fails validation because no attributes are being changed",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - a-test-key=a-test-value
  name: test-secret
  namespace: test-ns
  type: Opaque
`,
			Args:             []string{"test-secret", "--namespace=test-ns"},
			ExpectedErrorMsg: "at least one of [--from-literal, --new-namespace] must be specified",
		},
		{
			Name: "fails when a literal source doesn't have a key",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - some-key=some-value
  name: test-secret
  type: Opaque
`,
			Args:             []string{"test-secret", "--from-literal=value"},
			ExpectedErrorMsg: "literal values must be specified in the key=value format",
		},
		{
			Name: "fails when the secretGenerator field has no items",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator: []
`,
			Args:             []string{"test-secret", "--from-literal=value"},
			ExpectedErrorMsg: "unable to find Secret with name \"test-secret\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			kustomization, err := testutils_test.SetupEditSetConfigMapSecretTest(t, newCmdSetSecret, tc.KustomizationFileContent, tc.Args)

			require.Nil(t, kustomization)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.ExpectedErrorMsg)
		})
	}
}

func TestSuccessCasesEditSetSecret(t *testing.T) {
	testCases := []testutils_test.SuccessCase{
		{
			Name: "set a value successfully",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - random-key=random-value
  - key1=value
  name: test-secret
  type: Opaque
`,
			ExpectedLiterals: []string{"key1=val2", "random-key=random-value"},
			Args:             []string{"test-secret", "--from-literal=key1=val2"},
		},
		{
			Name: "successfully update namespace of target secret",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - a-key=test
  - another-key=value
  name: test-secret
  namespace: test-ns
  type: Opaque
`,
			Args:              []string{"test-secret", "--namespace=test-ns", "--new-namespace=test-new-ns"},
			ExpectedNamespace: "test-new-ns",
		},
		{
			Name: "successfully update namespace of target secret with empty namespace in file and namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - key1=value1
  - another-key=another-value
  name: test-secret
  type: Opaque
`,
			Args:              []string{"test-secret", "--namespace=default", "--new-namespace=test-new-ns"},
			ExpectedNamespace: "test-new-ns",
		},
		{
			Name: "successfully update namespace of target secret with default namespace and no namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - random-key=random-value
  name: test-secret
  namespace: default
  type: Opaque
`,
			Args:              []string{"test-secret", "--new-namespace=test-new-ns"},
			ExpectedNamespace: "test-new-ns",
		},
		{
			Name: "successfully update literal source of target secret with empty namespace in file and namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - a-key=a-value
  - key1=value
  name: test-secret
  type: Opaque
`,
			Args:             []string{"test-secret", "--namespace=default", "--from-literal=key1=val2"},
			ExpectedLiterals: []string{"a-key=a-value", "key1=val2"},
		},
		{
			Name: "successfully update namespace of target secret with default namespace and no namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - test-secret-key=value
  - key1=val1
  name: test-secret
  namespace: default
  type: Opaque
`,
			Args:             []string{"test-secret", "--from-literal=key1=val2"},
			ExpectedLiterals: []string{"test-secret-key=value", "key1=val2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			kustomization, err := testutils_test.SetupEditSetConfigMapSecretTest(t, newCmdSetSecret, tc.KustomizationFileContent, tc.Args)

			require.NoError(t, err)
			require.NotNil(t, kustomization)
			require.NotEmpty(t, kustomization.SecretGenerator)
			require.Greater(t, len(kustomization.SecretGenerator), 0)

			if tc.ExpectedNamespace != "" {
				require.Equal(t, tc.ExpectedNamespace, kustomization.SecretGenerator[0].Namespace)
			}

			if len(tc.ExpectedLiterals) > 0 {
				require.ElementsMatch(t, tc.ExpectedLiterals, kustomization.SecretGenerator[0].LiteralSources)
			}
		})
	}
}
