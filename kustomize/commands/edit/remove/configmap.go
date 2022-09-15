// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"fmt"
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
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	var newConfigMaps []types.ConfigMapArgs
	for _, configMap := range m.ConfigMapGenerator {
		if kustfile.StringInSlice(configMap.Name, o.configMapNames) {
			continue
		}
		newConfigMaps = append(newConfigMaps, configMap)
	}
	m.ConfigMapGenerator = newConfigMaps
	return mf.Write(m)
}
