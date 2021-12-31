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
	generatorFileName    = "myWonderfulGenerator.yaml"
	generatorFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddGeneratorHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(generatorFileName, []byte(generatorFileContent))
	require.NoError(t, err)
	err = fSys.WriteFile(generatorFileName+"another", []byte(generatorFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddGenerator(fSys)
	args := []string{generatorFileName + "*"}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	assert.Contains(t, string(content), generatorFileName)
	assert.Contains(t, string(content), generatorFileName+"another")
}

func TestAddGeneratorAlreadyThere(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(generatorFileName, []byte(generatorFileName))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddGenerator(fSys)
	args := []string{generatorFileName}
	assert.NoError(t, cmd.RunE(cmd, args))

	// adding an existing generator shouldn't return an error
	assert.NoError(t, cmd.RunE(cmd, args))

	// There can be only one. May it be the...
	mf, err := kustfile.NewKustomizationFile(fSys)
	assert.NoError(t, err)
	m, err := mf.Read()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.Generators))
	assert.Equal(t, generatorFileName, m.Generators[0])
}

func TestAddGeneratorNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddGenerator(fSys)
	err := cmd.Execute()
	assert.EqualError(t, err, "must specify a yaml file which contains a generator plugin resource")
}

func TestAddGeneratorMissingKustomizationYAML(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile(generatorFileName, []byte(generatorFileContent))
	require.NoError(t, err)
	err = fSys.WriteFile(generatorFileName+"another", []byte(generatorFileContent))
	require.NoError(t, err)

	cmd := newCmdAddGenerator(fSys)
	args := []string{generatorFileName + "*"}
	err = cmd.RunE(cmd, args)
	assert.EqualError(t, err, "Missing kustomization file 'kustomization.yaml'.\n")
}
