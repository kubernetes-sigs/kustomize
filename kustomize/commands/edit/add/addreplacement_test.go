// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	replaceFileName    = "replacement.yaml"
	replaceFileContent = `this is just a test file`
)

func TestAddReplacementWithFilePath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(replaceFileName, []byte(replaceFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddReplacement(fSys)
	args := []string{
		"--path", patchFileName,
	}
	cmd.SetArgs(args)
	assert.NoError(t, cmd.Execute())
	_, err = testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	kf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := kf.Read()
	require.NoError(t, err)

	expectedPath := []string{replaceFileName, patchFileName}

	for k, replacement := range kustomization.Replacements {
		require.Equal(t, expectedPath[k], replacement.Path)
	}
}

func TestAddReplacementAlreadyThere(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(replaceFileName, []byte(replaceFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddReplacement(fSys)
	args := []string{
		"--path", patchFileName,
	}
	cmd.SetArgs(args)
	assert.NoError(t, cmd.Execute())

	assert.Error(t, cmd.Execute())
}

func TestAddReplacementNoArgs(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()

	cmd := newCmdAddReplacement(fSys)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, "must provide path to add replacement", err.Error())
}
