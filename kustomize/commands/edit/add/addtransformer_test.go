// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	transformerFileName    = "myWonderfulTransformer.yaml"
	transformerFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddTransformerHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(transformerFileName, []byte(transformerFileContent))
	require.NoError(t, err)
	err = fSys.WriteFile(transformerFileName+"another", []byte(transformerFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddTransformer(fSys)
	args := []string{transformerFileName + "*"}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	assert.Contains(t, string(content), transformerFileName)
	assert.Contains(t, string(content), transformerFileName+"another")
}

func TestAddTransformerAlreadyThere(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(transformerFileName, []byte(transformerFileName))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddTransformer(fSys)
	args := []string{transformerFileName}
	assert.NoError(t, cmd.RunE(cmd, args))

	// adding an existing transformer shouldn't return an error
	assert.NoError(t, cmd.RunE(cmd, args))

	// There can be only one. May it be the...
	mf, err := kustfile.NewKustomizationFile(fSys)
	assert.NoError(t, err)
	m, err := mf.Read()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.Transformers))
	assert.Equal(t, transformerFileName, m.Transformers[0])
}

func TestAddTransformerNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddTransformer(fSys)
	err := cmd.Execute()
	assert.EqualError(t, err, "must specify a transformer file")
}

func TestAddTransformerMissingKustomizationYAML(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(transformerFileName, []byte(transformerFileContent))
	require.NoError(t, err)
	err = fSys.WriteFile(transformerFileName+"another", []byte(transformerFileContent))
	require.NoError(t, err)

	cmd := newCmdAddTransformer(fSys)
	args := []string{transformerFileName + "*"}
	err = cmd.RunE(cmd, args)
	assert.EqualError(t, err, "Missing kustomization file 'kustomization.yaml'.\n")
}
