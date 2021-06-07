// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
)

const (
	patchFileName    = "myWonderfulPatch.yaml"
	patchFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
`
	kind               = "myKind"
	group              = "myGroup"
	version            = "myVersion"
	name               = "myName"
	namespace          = "myNamespace"
	annotationSelector = "myAnnotationSelector"
	labelSelector      = "myLabelSelector"
)

func TestAddPatchWithFilePath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(patchFileName, []byte(patchFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{
		"--path", patchFileName,
		"--kind", kind,
		"--group", group,
		"--version", version,
		"--name", name,
		"--namespace", namespace,
		"--annotation-selector", annotationSelector,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	assert.NoError(t, cmd.Execute())
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	for i := 1; i < len(args); i += 2 {
		assert.Contains(t, string(content), args[i])
	}
}

func TestAddPatchWithPatchContent(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(patchFileName, []byte(patchFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{
		"--patch", patchFileContent,
		"--kind", kind,
		"--group", group,
		"--version", version,
		"--name", name,
		"--namespace", namespace,
		"--annotation-selector", annotationSelector,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	assert.NoError(t, cmd.Execute())
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	for i := 1; i < len(args); i += 2 {
		assert.Contains(t, string(content), strings.Trim(args[i], " \n"))
	}
}

func TestAddPatchAlreadyThere(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(patchFileName, []byte(patchFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{
		"--path", patchFileName,
		"--kind", kind,
		"--group", group,
		"--version", version,
		"--name", name,
		"--namespace", namespace,
		"--annotation-selector", annotationSelector,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	assert.NoError(t, cmd.Execute())

	// adding an existing patch shouldn't return an error
	assert.NoError(t, cmd.Execute())
}

func TestAddPatchNoArgs(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()

	cmd := newCmdAddPatch(fSys)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, "must provide either patch or path", err.Error())
}
