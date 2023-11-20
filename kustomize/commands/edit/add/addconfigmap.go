// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// newCmdAddConfigMap returns a new command.
func newCmdAddConfigMap(
	fSys filesys.FileSystem,
	ldr ifc.KvLoader,
	rf *resource.Factory,
) *cobra.Command {
	var flags util.ConfigMapSecretFlagsAndArgs
	cmd := &cobra.Command{
		Use:   "configmap NAME [--namespace=namespace-name] [--behavior={create|merge|replace}] [--from-file=[key=]source] [--from-literal=key1=value1]",
		Short: "Adds a configmap to the kustomization file",
		Long:  "",
		Example: `
	# Adds a configmap to the kustomization file (with a specified key)
	kustomize edit add configmap my-configmap --from-file=my-key=file/path --from-literal=my-literal=12345

	# Adds a configmap to the kustomization file (key is the filename)
	kustomize edit add configmap my-configmap --from-file=file/path

	# Adds a configmap from env-file
	kustomize edit add configmap my-configmap --from-env-file=env/path.env

	# Adds a configmap from env-file with behavior merge
	kustomize edit add configmap my-configmap --behavior=merge --from-env-file=env/path.env

	# Adds a configmap to the kustomization file with a specific namespace
	kustomize edit add configmap my-configmap --namespace test-ns --from-literal=my-key=my-value
`,
		RunE: func(_ *cobra.Command, args []string) error {
			return runEditAddConfigMap(flags, fSys, args, ldr, rf)
		},
	}

	cmd.Flags().StringSliceVar(
		&flags.FileSources,
		util.FromFileFlag,
		[]string{},
		"Key file can be specified using its file path, in which case file basename will be used as configmap "+
			"key, or optionally with a key and file path, in which case the given key will be used.  Specifying a "+
			"directory will iterate each named file in the directory whose basename is a valid configmap key.")
	cmd.Flags().StringArrayVar(
		&flags.LiteralSources,
		util.FromLiteralFlag,
		[]string{},
		"Specify a key and literal value to insert in configmap (i.e. mykey=somevalue)")
	cmd.Flags().StringVar(
		&flags.EnvFileSource,
		util.FromEnvFileFlag,
		"",
		"Specify the path to a file to read lines of key=val pairs to create a configmap (i.e. a Docker .env file).")
	cmd.Flags().BoolVar(
		&flags.DisableNameSuffixHash,
		util.DisableNameSuffixHashFlag,
		false,
		"Disable the name suffix for the configmap")
	cmd.Flags().StringVar(
		&flags.Behavior,
		util.BehaviorFlag,
		"",
		"Specify the behavior for config map generation, i.e whether to create a new configmap (the default),  "+
			"to merge with a previously defined one, or to replace an existing one. Merge and replace should be used only "+
			" when overriding an existing configmap defined in a base")
	cmd.Flags().StringVar(
		&flags.Namespace,
		util.NamespaceFlag,
		"",
		"Specify the namespace of the ConfigMap")

	return cmd
}

func runEditAddConfigMap(
	flags util.ConfigMapSecretFlagsAndArgs,
	fSys filesys.FileSystem,
	args []string,
	ldr ifc.KvLoader,
	rf *resource.Factory,
) error {
	err := flags.ExpandFileSource(fSys)
	if err != nil {
		return fmt.Errorf("failed to expand file source: %w", err)
	}

	err = flags.ValidateAdd(args)
	if err != nil {
		return fmt.Errorf("failed to validate flags: %w", err)
	}

	// Load the kustomization file.
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return fmt.Errorf("failed to load kustomization file: %w", err)
	}

	kustomization, err := mf.Read()
	if err != nil {
		return fmt.Errorf("failed to read kustomization file: %w", err)
	}

	// Add the configmapSecretFlagsAndArgs map to the kustomization file.
	err = addConfigMap(ldr, kustomization, flags, rf)
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}

	// Write out the kustomization file with added configmap.
	err = mf.Write(kustomization)
	if err != nil {
		return fmt.Errorf("failed to write kustomization file: %w", err)
	}

	return nil
}

// addConfigMap adds a configmap to a kustomization file.
// Note: error may leave kustomization file in an undefined state.
// Suggest passing a copy of kustomization file.
func addConfigMap(
	ldr ifc.KvLoader,
	k *types.Kustomization,
	flags util.ConfigMapSecretFlagsAndArgs,
	rf *resource.Factory,
) error {
	args := findOrMakeConfigMapArgs(k, flags.Name, flags.Namespace)
	util.MergeFlagsIntoGeneratorArgs(&args.GeneratorArgs, flags)
	// Validate by trying to create corev1.configmap.
	args.Options = types.MergeGlobalOptionsIntoLocal(
		args.Options, k.GeneratorOptions)
	_, err := rf.MakeConfigMap(ldr, args)
	return err
}

func findOrMakeConfigMapArgs(m *types.Kustomization, name, namespace string) *types.ConfigMapArgs {
	for i, v := range m.ConfigMapGenerator {
		if name == v.Name && util.NamespaceEqual(v.Namespace, namespace) {
			return &m.ConfigMapGenerator[i]
		}
	}
	// config map not found, create new one and add it to the kustomization file.
	cm := &types.ConfigMapArgs{
		GeneratorArgs: types.GeneratorArgs{Name: name, Namespace: namespace},
	}
	m.ConfigMapGenerator = append(m.ConfigMapGenerator, *cm)
	return &m.ConfigMapGenerator[len(m.ConfigMapGenerator)-1]
}
