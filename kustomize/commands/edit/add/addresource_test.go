// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
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
	err := fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	require.NoError(t, err)
	err = fSys.WriteFile(resourceFileName+"another", []byte(resourceFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName + "*"}
	assert.NoError(t, cmd.RunE(cmd, args))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	assert.Equal(t, string(content), `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: some-prefix
nameSuffix: some-suffix
# Labels to add to all objects and selectors.
# These labels would also be used to form the selector for apply --prune
# Named differently than “labels” to avoid confusion with metadata for this object
commonLabels:
  app: helloworld
commonAnnotations:
  note: This is an example annotation
resources:
- myWonderfulResource.yaml
- myWonderfulResource.yamlanother
#- service.yaml
#- ../some-dir/
# There could also be configmaps in Base, which would make these overlays
# There could be secrets in Base, if just using a fork/rebase workflow
replacements:
- path: replacement.yaml
  source: null
  targets: null
`)
}

func TestAddResourceAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName}
	assert.NoError(t, cmd.RunE(cmd, args))

	// adding an existing resource doesn't return an error
	assert.NoError(t, cmd.RunE(cmd, args))
}

func TestAddKustomizationFileAsResource(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	require.NoError(t, err)
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
