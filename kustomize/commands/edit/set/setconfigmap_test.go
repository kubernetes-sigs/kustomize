// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The duplicate linter is reporting that setsecret_test.go and this file are duplicates, which is not true.
// Disabling lint for these two files specifically to work around that.
//
//nolint:dupl
package set

import (
	"testing"

	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/configmapsecret"
)

func TestErrorCasesEditSetConfigMap(t *testing.T) {
	testCases := []testutils_test.FailureCase{
		{
			Name: "fails to set a value because key doesn't exist",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test-cm-key=value
  - key1=val1
  name: test-cm
- literals:
  - key3=val1
  name: test-cm-2
  namespace: test-ns
`,
			Args:             []string{"test-cm", "--from-literal=test-key=test-value"},
			ExpectedErrorMsg: "key 'test-key' not found in resource",
		},
		{
			Name: "fails to set a value because configmap doesn't exist",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - key1=val1
  - key2=val2
  name: test-cm
`,
			Args:             []string{"test-cm2", "--from-literal=test-key=test-value"},
			ExpectedErrorMsg: "unable to find ConfigMap with name \"test-cm2\"",
		},
		{
			Name: "fails validation because no attributes are being changed",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test-key=test-value
  name: test-cm
  namespace: test-ns
`,
			Args:             []string{"test-cm", "--namespace=test-ns"},
			ExpectedErrorMsg: "at least one of [--from-literal, --new-namespace] must be specified",
		},
		{
			Name: "fails when a literal source doesn't have a key",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test-key=test-value
  - key3=other-value
  name: test-cm
`,
			Args:             []string{"test-cm", "--from-literal=value"},
			ExpectedErrorMsg: "literal values must be specified in the key=value format",
		},
		{
			Name: "fails when the configMapGenerator field has no items",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator: []
`,
			Args:             []string{"test-cm", "--from-literal=value"},
			ExpectedErrorMsg: "unable to find ConfigMap with name \"test-cm\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			kustomization, err := testutils_test.SetupEditSetConfigMapSecretTest(t, newCmdSetConfigMap, tc.KustomizationFileContent, tc.Args)

			require.Nil(t, kustomization)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.ExpectedErrorMsg)
		})
	}
}

func TestSuccessCasesEditSetConfigMap(t *testing.T) {
	testCases := []testutils_test.SuccessCase{
		{
			Name: "set a value successfully",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - a-key=a-value
  - key1=val1
  name: test-cm
- literals:
  - another-key=another-value
  - key1=value-from-cm-2
  name: test-cm-2
  namespace: another-ns
`,
			ExpectedLiterals: []string{"a-key=a-value", "key1=val2"},
			Args:             []string{"test-cm", "--from-literal=key1=val2"},
		},
		{
			Name: "successfully update namespace of target configmap",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - yet-another-key=value
  name: test-cm
  namespace: test-ns
`,
			Args:              []string{"test-cm", "--namespace=test-ns", "--new-namespace=test-new-ns"},
			ExpectedNamespace: "test-new-ns",
		},
		{
			Name: "successfully update namespace of target configmap with empty namespace in file and namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - found-key=found-value
  - a-key=a-value
  name: test-cm
`,
			Args:              []string{"test-cm", "--namespace=default", "--new-namespace=test-new-ns"},
			ExpectedNamespace: "test-new-ns",
		},
		{
			Name: "successfully update namespace of target configmap with default namespace and no namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - a-different-key=a-different-value
  - key2=value
  name: test-cm
  namespace: default
`,
			Args:              []string{"test-cm", "--new-namespace=test-new-ns"},
			ExpectedNamespace: "test-new-ns",
		},
		{
			Name: "successfully update literal source of target configmap with empty namespace in file and namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - key1=value
  - a-separate-key=a-separate-value
  name: test-cm
`,
			Args:             []string{"test-cm", "--namespace=default", "--from-literal=key1=val2"},
			ExpectedLiterals: []string{"key1=val2", "a-separate-key=a-separate-value"},
		},
		{
			Name: "successfully update namespace of target configmap with default namespace and no namespace specified in command",
			KustomizationFileContent: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - a-random-key=a-random-value
  - key1=val1
  name: test-cm
  namespace: default
`,
			Args:             []string{"test-cm", "--from-literal=key1=val2"},
			ExpectedLiterals: []string{"a-random-key=a-random-value", "key1=val2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			kustomization, err := testutils_test.SetupEditSetConfigMapSecretTest(t, newCmdSetConfigMap, tc.KustomizationFileContent, tc.Args)

			require.NoError(t, err)
			require.NotNil(t, kustomization)
			require.NotEmpty(t, kustomization.ConfigMapGenerator)
			require.Greater(t, len(kustomization.ConfigMapGenerator), 0)

			if tc.ExpectedNamespace != "" {
				require.Equal(t, tc.ExpectedNamespace, kustomization.ConfigMapGenerator[0].Namespace)
			}

			if len(tc.ExpectedLiterals) > 0 {
				require.ElementsMatch(t, tc.ExpectedLiterals, kustomization.ConfigMapGenerator[0].LiteralSources)
			}
		})
	}
}
