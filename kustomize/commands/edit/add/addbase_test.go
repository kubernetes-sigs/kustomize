// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
)

const (
	baseDirectoryPaths = "my/path/to/wonderful/base,other/path/to/even/more/wonderful/base"
)

func TestAddBaseHappyPath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		fSys.Mkdir(base)
	}
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddBase(fSys)
	args := []string{baseDirectoryPaths}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	for _, base := range bases {
		assert.Contains(t, string(content), base)
	}
}

func TestAddBaseAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	// Create fake directories
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		fSys.Mkdir(base)
	}
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddBase(fSys)
	args := []string{baseDirectoryPaths}
	assert.NoError(t, cmd.RunE(cmd, args))
	// adding an existing base should return an error
	assert.Error(t, cmd.RunE(cmd, args))
	var expectedErrors []string
	for _, base := range bases {
		msg := "base " + base + " already in kustomization file"
		expectedErrors = append(expectedErrors, msg)
		assert.True(t, kustfile.StringInSlice(msg, expectedErrors))
	}
}

func TestAddBaseNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddBase(fSys)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, "must specify a base directory", err.Error())
}
