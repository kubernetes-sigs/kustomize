// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestNewCmdAddSecretIsNotNil(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	assert.NotNil(t, newCmdAddSecret(
		fSys,
		kv.NewLoader(
			loader.NewFileLoaderAtCwd(fSys),
			valtest_test.MakeFakeValidator()),
		nil))
}

func TestMakeSecretArgs(t *testing.T) {
	secretName := "test-secret-name"
	namespace := "test-secret-namespace"

	kustomization := &types.Kustomization{
		NamePrefix: "test-name-prefix",
	}

	secretType := "Opaque"

	assert.Equal(t, 0, len(kustomization.SecretGenerator))
	args := findOrMakeSecretArgs(kustomization, secretName, namespace, secretType)
	assert.NotNil(t, args)
	assert.Equal(t, 1, len(kustomization.SecretGenerator))
	assert.Equal(t, args, &kustomization.SecretGenerator[len(kustomization.SecretGenerator)-1])
	assert.Equal(t, args, findOrMakeSecretArgs(kustomization, secretName, namespace, secretType))
	assert.Equal(t, 1, len(kustomization.SecretGenerator))
}

func TestMergeFlagsIntoSecretArgs_LiteralSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{LiteralSources: []string{"k1=v1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{LiteralSources: []string{"k2=v2"}})
	assert.Equal(t, "k1=v1", k.SecretGenerator[0].LiteralSources[0])
	assert.Equal(t, "k2=v2", k.SecretGenerator[0].LiteralSources[1])
}

func TestMergeFlagsIntoSecretArgs_FileSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{FileSources: []string{"file1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{FileSources: []string{"file2"}})
	assert.Equal(t, "file1", k.SecretGenerator[0].FileSources[0])
	assert.Equal(t, "file2", k.SecretGenerator[0].FileSources[1])
}

func TestMergeFlagsIntoSecretArgs_EnvSource(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{EnvFileSource: "env1"})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{EnvFileSource: "env2"})
	assert.Equal(t, "env1", k.SecretGenerator[0].EnvSources[0])
	assert.Equal(t, "env2", k.SecretGenerator[0].EnvSources[1])
}

func TestMergeFlagsIntoSecretArgs_DisableNameSuffixHash(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{DisableNameSuffixHash: true})
	assert.True(t, k.SecretGenerator[0].Options.DisableNameSuffixHash)
}
