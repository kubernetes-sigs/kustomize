// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
)

const (
	resourceFileName    = "myWonderfulResource.yaml"
	resourceFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddResourceHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	fSys.WriteFile(resourceFileName+"another", []byte(resourceFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName + "*"}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	assert.Contains(t, string(content), resourceFileName)
	assert.Contains(t, string(content), resourceFileName+"another")
}

func TestAddResourceAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName}
	assert.NoError(t, cmd.RunE(cmd, args))

	// adding an existing resource doesn't return an error
	assert.NoError(t, cmd.RunE(cmd, args))
}

func TestAddKustomizationFileAsResource(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName}
	assert.NoError(t, cmd.RunE(cmd, args))

	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.NotContains(t, string(content), resourceFileName)
}

func TestAddResourceNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddResource(fSys)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, "must specify a resource file", err.Error())
}
