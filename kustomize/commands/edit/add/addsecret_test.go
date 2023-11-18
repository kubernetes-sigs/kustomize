// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
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
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{LiteralSources: []string{"k1=v1"}})
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{LiteralSources: []string{"k2=v2"}})
	assert.Equal(t, "k1=v1", k.SecretGenerator[0].LiteralSources[0])
	assert.Equal(t, "k2=v2", k.SecretGenerator[0].LiteralSources[1])
}

func TestMergeFlagsIntoSecretArgs_FileSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{FileSources: []string{"file1"}})
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{FileSources: []string{"file2"}})
	assert.Equal(t, "file1", k.SecretGenerator[0].FileSources[0])
	assert.Equal(t, "file2", k.SecretGenerator[0].FileSources[1])
}

func TestMergeFlagsIntoSecretArgs_EnvSource(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{EnvFileSource: "env1"})
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{EnvFileSource: "env2"})
	assert.Equal(t, "env1", k.SecretGenerator[0].EnvSources[0])
	assert.Equal(t, "env2", k.SecretGenerator[0].EnvSources[1])
}

func TestMergeFlagsIntoSecretArgs_DisableNameSuffixHash(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeSecretArgs(k, "foo", "bar", "forbidden")
	util.MergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		util.ConfigMapSecretFlagsAndArgs{DisableNameSuffixHash: true})
	assert.True(t, k.SecretGenerator[0].Options.DisableNameSuffixHash)
}

// TestEditAddSecretWithLiteralSource executes the same command flow as the CLI invocation
// with a --from-literal flag
func TestEditAddSecretWithLiteralSource(t *testing.T) {
	const (
		secretName    = "test-kustomization"
		literalSource = "test-key=test-value"
	)

	fSys := filesys.MakeEmptyDirInMemory()
	testutils_test.WriteTestKustomization(fSys)

	pvd := provider.NewDefaultDepProvider()
	ldr := kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), pvd.GetFieldValidator())

	args := []string{
		secretName,
		fmt.Sprintf(util.FlagFormat, util.FromLiteralFlag, literalSource),
	}
	cmd := newCmdAddSecret(fSys, ldr, pvd.GetResourceFactory())
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())

	_, err := testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := mf.Read()
	require.NoError(t, err)

	require.NotNil(t, kustomization)
	require.NotEmpty(t, kustomization.SecretGenerator)
	require.Equal(t, 1, len(kustomization.SecretGenerator))

	newSecretGenerator := kustomization.SecretGenerator[0]
	require.NotNil(t, newSecretGenerator)
	require.Equal(t, secretName, newSecretGenerator.Name)
	require.Contains(t, newSecretGenerator.LiteralSources, literalSource)
}

// TestEditAddSecretWithEnvSource executes the same command flow as the CLI invocation
// with a --from-env-file flag
func TestEditAddSecretWithEnvSource(t *testing.T) {
	const (
		secretName = "test-kustomization"
		envSource  = "test-env-source"
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
		secretName,
		fmt.Sprintf(util.FlagFormat, util.FromEnvFileFlag, envSource),
	}
	cmd := newCmdAddSecret(fSys, ldr, pvd.GetResourceFactory())
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())

	_, err = testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := mf.Read()
	require.NoError(t, err)

	require.NotNil(t, kustomization)
	require.NotEmpty(t, kustomization.SecretGenerator)
	require.Equal(t, 1, len(kustomization.SecretGenerator))

	newSecretGenerator := kustomization.SecretGenerator[0]
	require.NotNil(t, newSecretGenerator)
	require.Equal(t, secretName, newSecretGenerator.Name)
	require.Contains(t, newSecretGenerator.EnvSources, envSource)
}

// TestEditAddSecretWithFileSource executes the same command flow as the CLI invocation
// with a --from-file flag
func TestEditAddSecretWithFileSource(t *testing.T) {
	const (
		secretName = "test-kustomization"
		fileSource = "test-file-source"
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
		secretName,
		fmt.Sprintf(util.FlagFormat, util.FromFileFlag, fileSource),
	}
	cmd := newCmdAddSecret(fSys, ldr, pvd.GetResourceFactory())
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())

	_, err = testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	mf, err := kustfile.NewKustomizationFile(fSys)
	require.NoError(t, err)

	kustomization, err := mf.Read()
	require.NoError(t, err)

	require.NotNil(t, kustomization)
	require.NotEmpty(t, kustomization.SecretGenerator)
	require.Equal(t, 1, len(kustomization.SecretGenerator))

	newSecretGenerator := kustomization.SecretGenerator[0]
	require.NotNil(t, newSecretGenerator)
	require.Equal(t, secretName, newSecretGenerator.Name)
	require.Contains(t, newSecretGenerator.FileSources, fileSource)
}

// TestEditAddSecretNamespaced tests situations regarding namespacing. For example, it
// verifies that the empty namespace and the default namespace are treated the
// same when adding a configmap to a kustomization file.
func TestEditAddSecretNamespaced(t *testing.T) {
	testCases := []struct {
		name                string
		secretName          string
		secretNamespace     string
		literalSources      []string
		initialArgs         string
		expectedResult      []types.SecretArgs
		expectedSliceLength int
	}{
		{
			name:            "adds new key to secret when default namespace matches empty",
			secretName:      "test-secret",
			secretNamespace: "default",
			literalSources:  []string{"key1=value1"},
			initialArgs: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - key=value
  name: test-secret
  type: Opaque
`,
			expectedResult: []types.SecretArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "",
						Name:      "test-secret",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key=value", "key1=value1"},
						},
					},
					Type: ifc.SecretTypeOpaque,
				},
			},
			expectedSliceLength: 1,
		},
		{
			name:            "adds new key to secret when empty namespace matches default",
			secretName:      "test-secret",
			secretNamespace: "",
			literalSources:  []string{"key1=value1"},
			initialArgs: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - key=value
  name: test-secret
  namespace: default
  type: Opaque
`,
			expectedResult: []types.SecretArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "default",
						Name:      "test-secret",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key=value", "key1=value1"},
						},
					},
					Type: ifc.SecretTypeOpaque,
				},
			},
			expectedSliceLength: 1,
		},
		{
			name:            "creates a new generator when namespaces don't match",
			secretName:      "test-secret",
			secretNamespace: "",
			literalSources:  []string{"key1=value1"},
			initialArgs: `---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- literals:
  - key=value
  name: test-secret
  namespace: ns1
  type: Opaque
`,
			expectedResult: []types.SecretArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "ns1",
						Name:      "test-secret",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key=value"},
						},
					},
					Type: ifc.SecretTypeOpaque,
				},
				{
					GeneratorArgs: types.GeneratorArgs{
						Namespace: "",
						Name:      "test-secret",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"key1=value1"},
						},
					},
					Type: ifc.SecretTypeOpaque,
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
				tc.secretName,
				fmt.Sprintf(util.FlagFormat, util.NamespaceFlag, tc.secretNamespace),
			}

			for _, source := range tc.literalSources {
				args = append(args, fmt.Sprintf(util.FlagFormat, util.FromLiteralFlag, source))
			}

			cmd := newCmdAddSecret(fSys, ldr, pvd.GetResourceFactory())
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())

			_, err := testutils_test.ReadTestKustomization(fSys)
			require.NoError(t, err)

			mf, err := kustfile.NewKustomizationFile(fSys)
			require.NoError(t, err)

			kustomization, err := mf.Read()
			require.NoError(t, err)

			require.Len(t, kustomization.SecretGenerator, tc.expectedSliceLength)
			require.ElementsMatch(t, tc.expectedResult, kustomization.SecretGenerator)
		})
	}
}
