// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/pkg/loader"
	"sigs.k8s.io/kustomize/api/provider"
	. "sigs.k8s.io/kustomize/kustomize/v5/commands/edit/set"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestEditSetConfigMapError(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	pvd := provider.NewDefaultDepProvider()

	testCases := []struct {
		name             string
		input            string
		args             []string
		expectedErrorMsg string
	}{
		{
			name: "fail to set a value because key doesn't exist",
			input: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test=value
  - key1=val1
  name: test-cm
`,
			args:             []string{"test-cm", "--from-literal=key3=val2"},
			expectedErrorMsg: "key 'key3' not found in resource",
		},
		{
			name: "fail to set a value because configmap doesn't exist",
			input: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test=value
  - key1=val1
  name: test-cm
`,
			args:             []string{"test-cm2", "--from-literal=key3=val2"},
			expectedErrorMsg: "unable to find ConfigMap with name 'test-cm2'",
		},
		{
			name: "error on validate because no attributes are being changed",
			input: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test=value
  - key1=val1
  name: test-cm
  namespace: test-ns
`,
			args:             []string{"test-cm", "--namespace=test-ns"},
			expectedErrorMsg: "at least one of [--from-literal, --new-namespace] must be specified",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewCmdSetConfigMap(
				fSys,
				kv.NewLoader(
					loader.NewFileLoaderAtCwd(fSys),
					pvd.GetFieldValidator()),
				pvd.GetResourceFactory(),
			)

			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErrorMsg)
		})
	}
}

func TestEditSetConfigMapSuccess(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	pvd := provider.NewDefaultDepProvider()
	testCases := []struct {
		name              string
		input             string
		args              []string
		expectedLiterals  []string
		expectedNamespace string
	}{
		{
			name: "set a value successfully",
			input: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - key1=val1
  - test=value
  name: test-cm
`,
			expectedLiterals: []string{"key1=val2", "test=value"},
			args:             []string{"test-cm", "--from-literal=key1=val2"},
		},
		{
			name: "successfully update namespace of target configmap",
			input: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - test=value
  - key1=val1
  name: test-cm
  namespace: test-ns
`,
			args:              []string{"test-cm", "--namespace=test-ns", "--new-namespace=test-new-ns"},
			expectedNamespace: "test-new-ns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewCmdSetConfigMap(
				fSys,
				kv.NewLoader(
					loader.NewFileLoaderAtCwd(fSys),
					pvd.GetFieldValidator()),
				pvd.GetResourceFactory(),
			)

			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.input))

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			require.NoError(t, err)

			_, err = testutils_test.ReadTestKustomization(fSys)
			require.NoError(t, err)

			mf, err := kustfile.NewKustomizationFile(fSys)
			require.NoError(t, err)

			kustomization, err := mf.Read()
			require.NoError(t, err)

			require.NotNil(t, kustomization)
			require.NotEmpty(t, kustomization.ConfigMapGenerator)
			require.Greater(t, len(kustomization.ConfigMapGenerator), 0)

			if tc.expectedNamespace != "" {
				require.Equal(t, tc.expectedNamespace, kustomization.ConfigMapGenerator[0].Namespace)
			}

			if len(tc.expectedLiterals) > 0 {
				require.ElementsMatch(t, tc.expectedLiterals, kustomization.ConfigMapGenerator[0].LiteralSources)
			}
		})
	}
}
