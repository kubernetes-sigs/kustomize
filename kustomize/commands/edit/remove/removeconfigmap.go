// Copyright 2019 The Kubernetes Authors.
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
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeConfigMapOptions struct {
	configMapNamesToRemove []string
}

// newCmdRemoveResource remove the name of a file containing a resource to the kustomization file.
func newCmdRemoveConfigMap(fSys filesys.FileSystem) *cobra.Command {
	var o removeConfigMapOptions

	cmd := &cobra.Command{
		Use: "configmap",
		Short: "Removes specified configmap" +
			konfig.DefaultKustomizationFileName(),
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
	if len(args) == 0 {
		return errors.New("must specify a ConfigMap name")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: %s; to provide multiple ConfigMaps to remove, please separate ConfigMap names by commas", args)
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
		return fmt.Errorf("could not read kustomization file: %w", err)
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

	for _, name := range o.configMapNamesToRemove {
		if _, found := foundConfigMaps[name]; !found {
			log.Printf("configmap %s doesn't exist in kustomization file", name)
		}
	}

	m.ConfigMapGenerator = newConfigMaps
	err = mf.Write(m)
	if err != nil {
		return fmt.Errorf("configmap cannot write back to file, got %w", err)
	}
	return nil
}
