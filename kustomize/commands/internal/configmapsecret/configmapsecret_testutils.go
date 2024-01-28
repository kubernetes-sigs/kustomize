// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configmapsecret_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/pkg/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// FailureCase specifies a test case for a failure in the 'edit set configmap'/'edit set secret' commands.
type FailureCase struct {
	// The name of the test case
	Name string

	// The kustomization file content for the test case in YAML format
	KustomizationFileContent string

	// Arguments passed to the test case command
	Args []string

	// The expected error message for the test case
	ExpectedErrorMsg string
}

// SuccessCase specifies a test case for a success in the 'edit set configmap'/'edit set secret' commands.
type SuccessCase struct {
	// The name of the test case
	Name string

	// The kustomization file content for the test case in YAML format
	KustomizationFileContent string

	// Arguments passed to the test case command
	Args []string

	// List of expected literals for the result of the test case
	ExpectedLiterals []string

	// The expected namespace for the result of the test case
	ExpectedNamespace string
}

func SetupEditSetConfigMapSecretTest(
	t *testing.T,
	command func(filesys.FileSystem, ifc.KvLoader, *resource.Factory) *cobra.Command,
	input string,
	args []string,
) (*types.Kustomization, error) {
	t.Helper()
	fSys := filesys.MakeFsInMemory()
	pvd := provider.NewDefaultDepProvider()

	cmd := command(
		fSys,
		kv.NewLoader(
			loader.NewFileLoaderAtCwd(fSys),
			pvd.GetFieldValidator()),
		pvd.GetResourceFactory(),
	)

	testutils_test.WriteTestKustomizationWith(fSys, []byte(input))

	cmd.SetArgs(args)
	err := cmd.Execute()

	//nolint: wrapcheck
	// this needs to be bubbled up for checking in the test
	if err != nil {
		return nil, err
	}

	_, err = testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	//nolint: wrapcheck
	return mf.Read()
}
