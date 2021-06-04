// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
)

const (
	componentFileName    = "myWonderfulComponent.yaml"
	componentFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddComponentHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(componentFileName, []byte(componentFileContent))
	require.NoError(t, err)
	err = fSys.WriteFile(componentFileName+"another", []byte(componentFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddComponent(fSys)
	args := []string{componentFileName + "*"}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	assert.Contains(t, string(content), componentFileName)
	assert.Contains(t, string(content), componentFileName+"another")
}

func TestAddComponentAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile(componentFileName, []byte(componentFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddComponent(fSys)
	args := []string{componentFileName}
	assert.NoError(t, cmd.RunE(cmd, args))

	// adding an existing component doesn't return an error
	assert.NoError(t, cmd.RunE(cmd, args))
}

func TestAddKustomizationFileAsComponent(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile(componentFileName, []byte(componentFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddComponent(fSys)
	args := []string{componentFileName}
	require.NoError(t, cmd.RunE(cmd, args))

	content, err := testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)
	assert.NotContains(t, string(content), componentFileName)
}

func TestAddComponentNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddComponent(fSys)
	err := cmd.Execute()
	assert.EqualError(t, err, "must specify a component file")
}
