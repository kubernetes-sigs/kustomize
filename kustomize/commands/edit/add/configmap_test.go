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

const (
	configMapNamespace = "test-ns"
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
	args := findOrMakeConfigMapArgs(kustomization, cmName, configMapNamespace)
	assert.NotNil(t, args)
	assert.Equal(t, 1, len(kustomization.ConfigMapGenerator))
	assert.Equal(t, &kustomization.ConfigMapGenerator[len(kustomization.ConfigMapGenerator)-1], args)
	assert.Equal(t, args, findOrMakeConfigMapArgs(kustomization, cmName, configMapNamespace))
	assert.Equal(t, 1, len(kustomization.ConfigMapGenerator))
}

func TestMergeFlagsIntoConfigMapArgs_LiteralSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)
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
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)
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
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)
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
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)

	createBehaviorFlags := configmapSecretFlagsAndArgs{
		Behavior:      "create",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		createBehaviorFlags)
	require.Equal(t, configMapNamespace, k.ConfigMapGenerator[0].Namespace)
	assert.Equal(t, "create", k.ConfigMapGenerator[0].Behavior)

	mergeBehaviorFlags := configmapSecretFlagsAndArgs{
		Behavior:      "merge",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		mergeBehaviorFlags)
	require.Equal(t, configMapNamespace, k.ConfigMapGenerator[0].Namespace)
	assert.Equal(t, "merge", k.ConfigMapGenerator[0].Behavior)

	replaceBehaviorFlags := configmapSecretFlagsAndArgs{
		Behavior:      "replace",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		replaceBehaviorFlags)
	require.Equal(t, configMapNamespace, k.ConfigMapGenerator[0].Namespace)
	assert.Equal(t, "replace", k.ConfigMapGenerator[0].Behavior)
}

// TestEditAddConfigMapWithLiteralSource executes the same command flow as the CLI invocation
// with a --from-literal flag
func TestEditAddConfigMapWithLiteralSource(t *testing.T) {
	const (
		configMapName = "test-kustomization"
	)

	testCases := []struct {
		name               string
		literalSource      string
		configMapName      string
		configMapNamespace string
	}{
		{
			name:               "use literal-source with default namespace",
			literalSource:      "test-key=test-value",
			configMapName:      configMapName,
			configMapNamespace: "",
		},
		{
			name:               "use literal-source with specified namespace",
			literalSource:      "other-key=other-value",
			configMapName:      configMapName,
			configMapNamespace: configMapNamespace,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fSys := filesys.MakeEmptyDirInMemory()
			testutils_test.WriteTestKustomization(fSys)

			pvd := provider.NewDefaultDepProvider()
			ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

			args := []string{
				tc.configMapName,
				fmt.Sprintf(flagFormat, fromLiteralFlag, tc.literalSource),
				fmt.Sprintf(flagFormat, namespaceFlag, tc.configMapNamespace),
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
			require.Equal(t, tc.configMapName, newCmGenerator.Name)
			require.Equal(t, tc.configMapNamespace, newCmGenerator.Namespace)
			require.Contains(t, newCmGenerator.LiteralSources, tc.literalSource)
		})
	}
}

// TestEditAddConfigMapWithEnvSource executes the same command flow as the CLI invocation
// with a --from-env-file flag
func TestEditAddConfigMapWithEnvSource(t *testing.T) {
	const (
		configMapName = "test-kustomization"
		envSource     = "test-env-source"
		content       = "TEST=value"
	)

	testCases := []struct {
		name               string
		envSource          string
		content            string
		configMapName      string
		configMapNamespace string
	}{
		{
			name:               "use env-source with default namespace",
			envSource:          envSource,
			content:            content,
			configMapName:      configMapName,
			configMapNamespace: "",
		},
		{
			name: "use env-source with a specified namespace",

			envSource:          envSource,
			content:            content,
			configMapName:      configMapName,
			configMapNamespace: configMapNamespace,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fSys := filesys.MakeEmptyDirInMemory()
			testutils_test.WriteTestKustomization(fSys)

			pvd := provider.NewDefaultDepProvider()
			ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

			envFileContent, err := fSys.Create(tc.envSource)
			require.NoError(t, err)

			_, err = envFileContent.Write([]byte(tc.content))
			require.NoError(t, err)

			err = envFileContent.Close()
			require.NoError(t, err)

			args := []string{
				tc.configMapName,
				fmt.Sprintf(flagFormat, fromEnvFileFlag, tc.envSource),
				fmt.Sprintf(flagFormat, namespaceFlag, tc.configMapNamespace),
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
			require.Equal(t, tc.configMapName, newCmGenerator.Name)
			require.Equal(t, tc.configMapNamespace, newCmGenerator.Namespace)
			require.Contains(t, newCmGenerator.EnvSources, tc.envSource)
		})
	}
}

// TestEditAddConfigMapWithFileSource executes the same command flow as the CLI invocation
// with a --from-file flag
func TestEditAddConfigMapWithFileSource(t *testing.T) {
	const (
		configMapName = "test-kustomization"
		fileSource    = "test-file-source"
		content       = "any content here"
	)

	testCases := []struct {
		name               string
		fileSource         string
		content            string
		configMapName      string
		configMapNamespace string
	}{
		{
			name:               "use file-source with default namespace",
			fileSource:         fileSource,
			content:            content,
			configMapName:      configMapName,
			configMapNamespace: "",
		},
		{
			name:               "use file-source with specified namespace",
			fileSource:         fileSource,
			content:            content,
			configMapName:      configMapName,
			configMapNamespace: configMapNamespace,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fSys := filesys.MakeEmptyDirInMemory()
			testutils_test.WriteTestKustomization(fSys)

			pvd := provider.NewDefaultDepProvider()
			ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

			fileContent, err := fSys.Create(tc.fileSource)
			require.NoError(t, err)

			_, err = fileContent.Write([]byte(tc.content))
			require.NoError(t, err)

			err = fileContent.Close()
			require.NoError(t, err)

			args := []string{
				tc.configMapName,
				fmt.Sprintf(flagFormat, fromFileFlag, tc.fileSource),
				fmt.Sprintf(flagFormat, namespaceFlag, tc.configMapNamespace),
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
			require.Equal(t, tc.configMapName, newCmGenerator.Name)
			require.Equal(t, tc.configMapNamespace, newCmGenerator.Namespace)
			require.Contains(t, newCmGenerator.FileSources, tc.fileSource)
		})
	}
}
