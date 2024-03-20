// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
//nolint:dupl
package set

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/konfig"

	"slices"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func newCmdSetConfigMap(
	fSys filesys.FileSystem,
	ldr ifc.KvLoader,
	rf *resource.Factory,
) *cobra.Command {
	var flags util.ConfigMapSecretFlagsAndArgs
	cmd := &cobra.Command{
		Use:   "configmap NAME [--from-literal=key1=value1] [--namespace=namespace-name] [--new-namespace=new-namespace-name]",
		Short: fmt.Sprintf("Edits the value for an existing key for a ConfigMap in the %s file", konfig.DefaultKustomizationFileName()),
		Long: fmt.Sprintf(`Edits the value for an existing key in an existing ConfigMap in the %[1]s file.
ConfigMap name, ConfigMap namespace, and key name must match an existing entry in the %[1]s file for this command to succeed.
When namespace is omitted, the default namespace is used. Conversely, when an entry without a specified namespace exists
in the %[1]s file, it can be updated by either omitting the namespace on the kustomize edit set configmap invocation or by
specifying --namespace=default.`, konfig.DefaultKustomizationFileName()),
		Example: fmt.Sprintf(`
	# Edits an existing ConfigMap in the %[1]s file, changing value of key1 to 2, and namespace is implicitly defined as "default"
	kustomize edit set configmap my-configmap --from-literal=key1=2

	# Edits an existing ConfigMap in the %[1]s file, changing value of key1 to 2, and explicitly define namespace as "default"
	kustomize edit set configmap my-configmap --from-literal=key1=2 --namespace default

	# Edits an existing ConfigMap in the %[1]s file, changing namespace to "new-namespace"
	kustomize edit set configmap my-configmap --namespace=current-namespace --new-namespace=new-namespace
`, konfig.DefaultKustomizationFileName()),
		RunE: func(_ *cobra.Command, args []string) error {
			return runEditSetConfigMap(flags, fSys, args, ldr, rf)
		},
	}

	cmd.Flags().StringArrayVar(
		&flags.LiteralSources,
		util.FromLiteralFlag,
		[]string{},
		"Specify an existing key and a new value to update a ConfigMap (i.e. mykey=newvalue)")
	cmd.Flags().StringVar(
		&flags.Namespace,
		util.NamespaceFlag,
		"",
		"Current namespace of the target ConfigMap")
	cmd.Flags().StringVar(
		&flags.NewNamespace,
		util.NewNamespaceFlag,
		"",
		"New namespace value for the target ConfigMap")

	return cmd
}

func runEditSetConfigMap(
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

	err = flags.ValidateSet(args)
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

	// Updates the existing ConfigMap
	err = setConfigMap(ldr, kustomization, flags, rf)
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

func setConfigMap(
	ldr ifc.KvLoader,
	k *types.Kustomization,
	flags util.ConfigMapSecretFlagsAndArgs,
	rf *resource.Factory,
) error {
	args, err := findConfigMapArgs(k, flags.Name, flags.Namespace)
	if err != nil {
		return fmt.Errorf("could not set new ConfigMap value: %w", err)
	}

	if len(flags.LiteralSources) > 0 {
		err := util.UpdateLiteralSources(&args.GeneratorArgs, flags)
		if err != nil {
			return fmt.Errorf("failed to update literal sources: %w", err)
		}
	}

	// update namespace to new one
	if flags.NewNamespace != "" {
		args.Namespace = flags.NewNamespace
	}

	// Validate by trying to create corev1.configmap.
	args.Options = types.MergeGlobalOptionsIntoLocal(
		args.Options, k.GeneratorOptions)

	_, err = rf.MakeConfigMap(ldr, args)
	if err != nil {
		return fmt.Errorf("failed to validate ConfigMap structure: %w", err)
	}

	return nil
}

// findConfigMapArgs finds the generator arguments corresponding to the specified
// ConfigMap name. ConfigMap must exist for this command to be successful.
func findConfigMapArgs(m *types.Kustomization, name, namespace string) (*types.ConfigMapArgs, error) {
	cmIndex := slices.IndexFunc(m.ConfigMapGenerator, func(cmArgs types.ConfigMapArgs) bool {
		return name == cmArgs.Name && util.NamespaceEqual(namespace, cmArgs.Namespace)
	})

	if cmIndex == -1 {
		return nil, fmt.Errorf("unable to find ConfigMap with name %q", name)
	}

	return &m.ConfigMapGenerator[cmIndex], nil
}
