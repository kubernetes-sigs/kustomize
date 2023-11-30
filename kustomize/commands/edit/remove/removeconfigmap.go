// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeConfigMapOptions struct {
	configMapNamesToRemove []string
	namespace              string
}

// newCmdRemoveConfigMap removes configMapGenerator(s) with the specified name(s).
func newCmdRemoveConfigMap(fSys filesys.FileSystem) *cobra.Command {
	var flags removeConfigMapOptions

	cmd := &cobra.Command{
		Use: "configmap NAME [,NAME] [--namespace=namespace-name]",
		Short: "Removes the specified configmap(s) from " +
			konfig.DefaultKustomizationFileName(),
		Long: `Removes the specified configmap(s) from the ` + konfig.DefaultKustomizationFileName() + ` file in the specified namespace.
If multiple configmap names are specified, the command will not fail on secret names that were not found in the file,
but will issue a warning for each name that wasn't found.`,
		Example: `
    # Removes a single configmap named 'my-configmap' in the default namespace from the ` + konfig.DefaultKustomizationFileName() + ` file
    kustomize edit remove configmap my-configmap

    # Removes configmaps named 'my-configmap' and 'other-configmap' in namespace 'test-namespace' from the ` + konfig.DefaultKustomizationFileName() + ` file
    kustomize edit remove configmap my-configmap,other-configmap --namespace=test-namespace
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := flags.Validate(args)
			if err != nil {
				return err
			}
			return flags.RunRemoveConfigMap(fSys)
		},
	}

	cmd.Flags().StringVar(
		&flags.namespace,
		util.NamespaceFlag,
		"",
		"Namespace to remove ConfigMap(s) from",
	)

	return cmd
}

// Validate validates removeConfigMap command.
func (o *removeConfigMapOptions) Validate(args []string) error {
	switch {
	case len(args) == 0:
		return errors.New("at least one configmap name must be specified")
	case len(args) > 1:
		return fmt.Errorf("too many arguments: %s; to provide multiple configmaps to remove, please separate configmap names by commas", args)
	}

	o.configMapNamesToRemove = strings.Split(args[0], ",")
	return nil
}

// RunRemoveConfigMap runs ConfigMap command (do real work).
func (o *removeConfigMapOptions) RunRemoveConfigMap(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return fmt.Errorf("could not read kustomization file: %w", err)
	}

	m, err := mf.Read()
	if err != nil {
		return fmt.Errorf("could not read kustomization file contents: %w", err)
	}

	foundConfigMaps := make(map[string]struct{})
	remainingConfigMaps := make([]types.ConfigMapArgs, 0, len(m.ConfigMapGenerator))

	for _, currentConfigMap := range m.ConfigMapGenerator {
		if kustfile.StringInSlice(currentConfigMap.Name, o.configMapNamesToRemove) &&
			util.NamespaceEqual(currentConfigMap.Namespace, o.namespace) {
			foundConfigMaps[currentConfigMap.Name] = struct{}{}
			continue
		}

		remainingConfigMaps = append(remainingConfigMaps, currentConfigMap)
	}

	if len(foundConfigMaps) == 0 {
		return fmt.Errorf("no specified configmap(s) were found in the %s file",
			konfig.DefaultKustomizationFileName())
	}

	for _, name := range o.configMapNamesToRemove {
		if _, found := foundConfigMaps[name]; !found {
			log.Printf("configmap %s doesn't exist in kustomization file", name)
		}
	}

	m.ConfigMapGenerator = remainingConfigMaps
	err = mf.Write(m)
	if err != nil {
		return fmt.Errorf("failed to write kustomization file: %w", err)
	}
	return nil
}
