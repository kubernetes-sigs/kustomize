// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/pkg/loader"
	"sigs.k8s.io/kustomize/api/provider"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
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
		configmapSecretFlagsAndArgs{LiteralSources: []string{"k1=v1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		configmapSecretFlagsAndArgs{LiteralSources: []string{"k2=v2"}})
	assert.Equal(t, "k1=v1", k.ConfigMapGenerator[0].LiteralSources[0])
	assert.Equal(t, "k2=v2", k.ConfigMapGenerator[0].LiteralSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_FileSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		configmapSecretFlagsAndArgs{FileSources: []string{"file1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		configmapSecretFlagsAndArgs{FileSources: []string{"file2"}})
	assert.Equal(t, "file1", k.ConfigMapGenerator[0].FileSources[0])
	assert.Equal(t, "file2", k.ConfigMapGenerator[0].FileSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_EnvSource(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		configmapSecretFlagsAndArgs{EnvFileSource: "env1"})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		configmapSecretFlagsAndArgs{EnvFileSource: "env2"})
	assert.Equal(t, "env1", k.ConfigMapGenerator[0].EnvSources[0])
	assert.Equal(t, "env2", k.ConfigMapGenerator[0].EnvSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_Behavior(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")

	createBehaviorFlags := configmapSecretFlagsAndArgs{
		Behavior:      "create",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		createBehaviorFlags)
	assert.Equal(t, "create", k.ConfigMapGenerator[0].Behavior)

	mergeBehaviorFlags := configmapSecretFlagsAndArgs{
		Behavior:      "merge",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		mergeBehaviorFlags)
	assert.Equal(t, "merge", k.ConfigMapGenerator[0].Behavior)

	replaceBehaviorFlags := configmapSecretFlagsAndArgs{
		Behavior:      "replace",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		replaceBehaviorFlags)
	assert.Equal(t, "replace", k.ConfigMapGenerator[0].Behavior)
}

// TestEditAddConfigMapWithLiteralSource executes the same command flow as the CLI invocation
// with a --from-literal flag
func TestEditAddConfigMapWithLiteralSource(t *testing.T) {
	const (
		configMapName = "test-kustomization"
		literalSource = "test-key=test-value"
	)

	fSys := filesys.MakeEmptyDirInMemory()
	testutils_test.WriteTestKustomization(fSys)

	pvd := provider.NewDefaultDepProvider()
	ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

	args := []string{
		configMapName,
		fmt.Sprintf(flagFormat, fromLiteralFlag, literalSource),
	}
	cmd := newCmdAddConfigMap(fSys, ldr, pvd.GetResourceFactory())
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())

	_, err := testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := mf.Read()
	require.NoError(t, err)

	require.NotNil(t, kustomization)
	require.NotEmpty(t, kustomization.ConfigMapGenerator)
	require.Equal(t, 1, len(kustomization.ConfigMapGenerator))

	newCmGenerator := kustomization.ConfigMapGenerator[0]
	require.NotNil(t, newCmGenerator)
	require.Equal(t, configMapName, newCmGenerator.Name)
	require.Contains(t, newCmGenerator.LiteralSources, literalSource)
}

// TestEditAddConfigMapWithEnvSource executes the same command flow as the CLI invocation
// with a --from-env-file flag
func TestEditAddConfigMapWithEnvSource(t *testing.T) {
	const (
		configMapName = "test-kustomization"
		envSource     = "test-env-source"
	)

	fSys := filesys.MakeEmptyDirInMemory()
	testutils_test.WriteTestKustomization(fSys)

	pvd := provider.NewDefaultDepProvider()
	ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

	envFileContent, err := fSys.Create("test-env-source")
	require.NoError(t, err)

	_, err = envFileContent.Write([]byte("TEST=value"))
	require.NoError(t, err)

	err = envFileContent.Close()
	require.NoError(t, err)

	args := []string{
		configMapName,
		fmt.Sprintf(flagFormat, fromEnvFileFlag, envSource),
	}
	cmd := newCmdAddConfigMap(fSys, ldr, pvd.GetResourceFactory())
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())

	_, err = testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := mf.Read()
	require.NoError(t, err)

	require.NotNil(t, kustomization)
	require.NotEmpty(t, kustomization.ConfigMapGenerator)
	require.Equal(t, 1, len(kustomization.ConfigMapGenerator))

	newCmGenerator := kustomization.ConfigMapGenerator[0]
	require.NotNil(t, newCmGenerator)
	require.Equal(t, configMapName, newCmGenerator.Name)
	require.Contains(t, newCmGenerator.EnvSources, envSource)
}

// TestEditAddConfigMapWithFileSource executes the same command flow as the CLI invocation
// with a --from-file flag
func TestEditAddConfigMapWithFileSource(t *testing.T) {
	const (
		configMapName = "test-kustomization"
		fileSource    = "test-file-source"
	)

	fSys := filesys.MakeEmptyDirInMemory()
	testutils_test.WriteTestKustomization(fSys)

	pvd := provider.NewDefaultDepProvider()
	ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

	fileContent, err := fSys.Create("test-file-source")
	require.NoError(t, err)

	_, err = fileContent.Write([]byte("any content here"))
	require.NoError(t, err)

	err = fileContent.Close()
	require.NoError(t, err)

	args := []string{
		configMapName,
		fmt.Sprintf(flagFormat, fromFileFlag, fileSource),
	}
	cmd := newCmdAddConfigMap(fSys, ldr, pvd.GetResourceFactory())
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())

	_, err = testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := mf.Read()
	require.NoError(t, err)

	require.NotNil(t, kustomization)
	require.NotEmpty(t, kustomization.ConfigMapGenerator)
	require.Equal(t, 1, len(kustomization.ConfigMapGenerator))

	newCmGenerator := kustomization.ConfigMapGenerator[0]
	require.NotNil(t, newCmGenerator)
	require.Equal(t, configMapName, newCmGenerator.Name)
	require.Contains(t, newCmGenerator.FileSources, fileSource)
}
