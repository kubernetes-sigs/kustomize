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
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeConfigMapOptions struct {
	configMapNamesToRemove []string
}

// newCmdRemoveConfigMap removes configMapGenerator(s) with the specified name(s).
func newCmdRemoveConfigMap(fSys filesys.FileSystem) *cobra.Command {
	var o removeConfigMapOptions

	cmd := &cobra.Command{
		Use: "configmap",
		Short: "Removes the specified configmap(s) from " +
			konfig.DefaultKustomizationFileName(),
		Long: "",
		Example: `
		remove configmap my-configmap
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunRemoveConfigMap(fSys)
		},
	}
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

	newConfigMaps := make([]types.ConfigMapArgs, 0, len(m.ConfigMapGenerator))
	for _, currentConfigMap := range m.ConfigMapGenerator {
		if kustfile.StringInSlice(currentConfigMap.Name, o.configMapNamesToRemove) {
			foundConfigMaps[currentConfigMap.Name] = struct{}{}
			continue
		}
		newConfigMaps = append(newConfigMaps, currentConfigMap)
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

	m.ConfigMapGenerator = newConfigMaps
	err = mf.Write(m)
	if err != nil {
		return fmt.Errorf("failed to write kustomization file: %w", err)
	}
	return nil
}
