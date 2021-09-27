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

func TestNewAddConfigMapIsNotNil(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	assert.NotNil(t, newCmdAddConfigMap(
		fSys,
		kv.NewLoader(
			loader.NewFileLoaderAtCwd(fSys),
			valtest_test.MakeFakeValidator()),
		nil))
}

func TestMakeConfigMapArgs(t *testing.T) {
	cmName := "test-config-name"

	kustomization := &types.Kustomization{
		NamePrefix: "test-name-prefix",
	}

	if len(kustomization.ConfigMapGenerator) != 0 {
		t.Fatal("Initial kustomization should not have any configmaps")
	}
	args := findOrMakeConfigMapArgs(kustomization, cmName)
	assert.NotNil(t, args)
	assert.Equal(t, 1, len(kustomization.ConfigMapGenerator))
	assert.Equal(t, &kustomization.ConfigMapGenerator[len(kustomization.ConfigMapGenerator)-1], args)
	assert.Equal(t, args, findOrMakeConfigMapArgs(kustomization, cmName))
	assert.Equal(t, 1, len(kustomization.ConfigMapGenerator))
}

func TestMergeFlagsIntoConfigMapArgs_LiteralSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{LiteralSources: []string{"k1=v1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{LiteralSources: []string{"k2=v2"}})
	assert.Equal(t, "k1=v1", k.ConfigMapGenerator[0].LiteralSources[0])
	assert.Equal(t, "k2=v2", k.ConfigMapGenerator[0].LiteralSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_FileSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{FileSources: []string{"file1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{FileSources: []string{"file2"}})
	assert.Equal(t, "file1", k.ConfigMapGenerator[0].FileSources[0])
	assert.Equal(t, "file2", k.ConfigMapGenerator[0].FileSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_EnvSource(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{EnvFileSource: "env1"})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{EnvFileSource: "env2"})
	assert.Equal(t, "env1", k.ConfigMapGenerator[0].EnvSources[0])
	assert.Equal(t, "env2", k.ConfigMapGenerator[0].EnvSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_Behavior(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")

	createBehaviorFlags := flagsAndArgs{
		Behavior:      "create",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		createBehaviorFlags)
	assert.Equal(t, "create", k.ConfigMapGenerator[0].Behavior)

	mergeBehaviorFlags := flagsAndArgs{
		Behavior:      "merge",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		mergeBehaviorFlags)
	assert.Equal(t, "merge", k.ConfigMapGenerator[0].Behavior)

	replaceBehaviorFlags := flagsAndArgs{
		Behavior:      "replace",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		replaceBehaviorFlags)
	assert.Equal(t, "replace", k.ConfigMapGenerator[0].Behavior)
}
