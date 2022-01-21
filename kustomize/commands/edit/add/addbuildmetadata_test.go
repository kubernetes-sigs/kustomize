// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/types"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestAddBuildMetadataHappyPath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	cmd := newCmdAddBuildMetadata(fSys)
	args := []string{strings.Join(types.BuildMetadataOptions, ",")}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	for _, opt := range types.BuildMetadataOptions {
		assert.Contains(t, string(content), opt)
	}
}

func TestAddBuildMetadataAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	cmd := newCmdAddBuildMetadata(fSys)
	args := []string{strings.Join(types.BuildMetadataOptions, ",")}
	assert.NoError(t, cmd.RunE(cmd, args))
	err := cmd.RunE(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in kustomization file")
}

func TestAddBuildMetadataInvalidArg(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	cmd := newCmdAddBuildMetadata(fSys)
	args := []string{"invalid_option"}
	err := cmd.RunE(cmd, args)
	assert.Error(t, err)
	assert.Equal(t, "invalid buildMetadata option: invalid_option", err.Error())
}

func TestAddBuildMetadataTooManyArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	cmd := newCmdAddBuildMetadata(fSys)
	args := []string{"option1", "option2"}
	err := cmd.RunE(cmd, args)
	assert.Error(t, err)
	assert.Equal(t, "too many arguments: [option1 option2]; to provide multiple buildMetadata options, please separate options by comma", err.Error())
}
