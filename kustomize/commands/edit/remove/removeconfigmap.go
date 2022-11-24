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
	configMapNames []string
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
		return errors.New("must specify a configmap name")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: %s; to provide multiple config map options, please separate options by comma", args)
	}
	o.configMapNames = strings.Split(args[0], ",")
	return nil
}

// RunRemoveConfigMap runs ConfigMap command (do real work).
func (o *removeConfigMapOptions) RunRemoveConfigMap(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return fmt.Errorf("configmap cannot load from file system, got %w", err)
	}

	m, err := mf.Read()
	if err != nil {
		return fmt.Errorf("configmap cannot read from file, got %w", err)
	}

	foundConfigMaps := make(map[string]bool)
	for _, removeName := range o.configMapNames {
		foundConfigMaps[removeName] = false
	}

	newConfigMaps := make([]types.ConfigMapArgs, 0, len(m.ConfigMapGenerator))
	for _, currentConfigMap := range m.ConfigMapGenerator {
		if kustfile.StringInSlice(currentConfigMap.Name, o.configMapNames) {
			foundConfigMaps[currentConfigMap.Name] = true
			continue
		}
		newConfigMaps = append(newConfigMaps, currentConfigMap)
	}

	for name, found := range foundConfigMaps {
		if !found {
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
