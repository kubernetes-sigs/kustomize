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
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	configMapNamespace = "test-ns"
)

func TestNewAddConfigMapIsNotNil(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	require.NotNil(t, newCmdAddConfigMap(
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

	require.Len(t, kustomization.ConfigMapGenerator, 0, "Initial kustomization should not have any configmaps")

	args := findOrMakeConfigMapArgs(kustomization, cmName, configMapNamespace)
	require.NotNil(t, args)
	require.Equal(t, 1, len(kustomization.ConfigMapGenerator))
	require.Equal(t, &kustomization.ConfigMapGenerator[len(kustomization.ConfigMapGenerator)-1], args)
	require.Equal(t, args, findOrMakeConfigMapArgs(kustomization, cmName, configMapNamespace))
	require.Equal(t, 1, len(kustomization.ConfigMapGenerator))
}

func TestMergeFlagsIntoConfigMapArgs_LiteralSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{LiteralSources: []string{"k1=v1"}})
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{LiteralSources: []string{"k2=v2"}})
	assert.Equal(t, "k1=v1", k.ConfigMapGenerator[0].LiteralSources[0])
	assert.Equal(t, "k2=v2", k.ConfigMapGenerator[0].LiteralSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_FileSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{FileSources: []string{"file1"}})
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{FileSources: []string{"file2"}})
	assert.Equal(t, "file1", k.ConfigMapGenerator[0].FileSources[0])
	assert.Equal(t, "file2", k.ConfigMapGenerator[0].FileSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_EnvSource(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{EnvFileSource: "env1"})
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{EnvFileSource: "env2"})
	assert.Equal(t, "env1", k.ConfigMapGenerator[0].EnvSources[0])
	assert.Equal(t, "env2", k.ConfigMapGenerator[0].EnvSources[1])
}

func TestMergeFlagsIntoConfigMapArgs_Behavior(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo", configMapNamespace)

	createBehaviorFlags := util.ConfigMapSecretFlagsAndArgs{
		Behavior:      "create",
		EnvFileSource: "env1",
	}
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		createBehaviorFlags)
	require.Equal(t, configMapNamespace, k.ConfigMapGenerator[0].Namespace)
	assert.Equal(t, "create", k.ConfigMapGenerator[0].Behavior)

	mergeBehaviorFlags := util.ConfigMapSecretFlagsAndArgs{
		Behavior:      "merge",
		EnvFileSource: "env1",
	}
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		mergeBehaviorFlags)
	require.Equal(t, configMapNamespace, k.ConfigMapGenerator[0].Namespace)
	assert.Equal(t, "merge", k.ConfigMapGenerator[0].Behavior)

	replaceBehaviorFlags := util.ConfigMapSecretFlagsAndArgs{
		Behavior:      "replace",
		EnvFileSource: "env1",
	}
	util.MergeFlagsIntoGeneratorArgs(
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
				fmt.Sprintf(util.FlagFormat, util.FromLiteralFlag, tc.literalSource),
				fmt.Sprintf(util.FlagFormat, util.NamespaceFlag, tc.configMapNamespace),
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
				fmt.Sprintf(util.FlagFormat, util.FromEnvFileFlag, tc.envSource),
				fmt.Sprintf(util.FlagFormat, util.NamespaceFlag, tc.configMapNamespace),
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
				fmt.Sprintf(util.FlagFormat, util.FromFileFlag, tc.fileSource),
				fmt.Sprintf(util.FlagFormat, util.NamespaceFlag, tc.configMapNamespace),
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

// TestEditAddConfigMapNamespaced tests situations regarding namespacing. For example, it
// verifies that the empty namespace and the default namespace are treated the
// same when adding a configmap to a kustomization file.
func TestEditAddConfigMapNamespaced(t *testing.T) {
	testCases := []struct {
		name                string
		configMapName       string
		configMapNamespace  string
		literalSources      []string
		initialArgs         string
		expectedResult      []types.ConfigMapArgs
		expectedSliceLength int
	}{
		{
			name:               "adds new key to configmap when default namespace matches empty",
			configMapName:      "test-cm",
			configMapNamespace: "default",
			literalSources:     []string{"key1=value1"},
			initialArgs: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - key=value
  name: test-cm
`,
			expectedResult: []types.ConfigMapArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "",
						Name:      "test-cm",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key=value", "key1=value1"},
						},
					},
				},
			},
			expectedSliceLength: 1,
		},
		{
			name:               "adds new key to configmap when empty namespace matches default",
			configMapName:      "test-cm",
			configMapNamespace: "",
			literalSources:     []string{"key1=value1"},
			initialArgs: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - key=value
  name: test-cm
  namespace: default
`,
			expectedResult: []types.ConfigMapArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "default",
						Name:      "test-cm",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key=value", "key1=value1"},
						},
					},
				},
			},
			expectedSliceLength: 1,
		},
		{
			name:               "creates a new generator when namespaces don't match",
			configMapName:      "test-cm",
			configMapNamespace: "",
			literalSources:     []string{"key1=value1"},
			initialArgs: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- literals:
  - key=value
  name: test-cm
  namespace: ns1
`,
			expectedResult: []types.ConfigMapArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "ns1",
						Name:      "test-cm",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key=value"},
						},
					},
				},
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "",
						Name:      "test-cm",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key1=value1"},
						},
					},
				},
			},
			expectedSliceLength: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fSys := filesys.MakeEmptyDirInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(tc.initialArgs))

			pvd := provider.NewDefaultDepProvider()
			ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

			args := []string{
				tc.configMapName,
				fmt.Sprintf(util.FlagFormat, util.NamespaceFlag, tc.configMapNamespace),
			}

			for _, source := range tc.literalSources {
				args = append(args, fmt.Sprintf(util.FlagFormat, util.FromLiteralFlag, source))
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

			require.Len(t, kustomization.ConfigMapGenerator, tc.expectedSliceLength)
			require.ElementsMatch(t, tc.expectedResult, kustomization.ConfigMapGenerator)
		})
	}
}
