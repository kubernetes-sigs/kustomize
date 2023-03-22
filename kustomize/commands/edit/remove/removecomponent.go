// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeComponentOptions struct {
	componentFilePaths []string
}

// newCmdRemoveComponent remove the name of a file containing a component to the kustomization file.
func newCmdRemoveComponent(fSys filesys.FileSystem) *cobra.Command {
	var o removeComponentOptions

	cmd := &cobra.Command{
		Use: "component",
		Short: "Removes one or more components from components field" +
			konfig.DefaultKustomizationFileName(),
		Example: `
		remove component ../../components/component1
		remove component ../../components/component1 ../../components/component2
		remove component ../../components/component*
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunRemoveComponent(fSys)
		},
	}
	return cmd
}

// Validate validates removeComponent command.
func (o *removeComponentOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a component")
	}
	o.componentFilePaths = args
	return nil
}

// RunRemoveComponent runs Component command (do real work).
func (o *removeComponentOptions) RunRemoveComponent(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	components, err := globPatterns(m.Components, o.componentFilePaths)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		return nil
	}

	newComponents := make([]string, 0, len(m.Components))
	for _, component := range m.Components {
		if kustfile.StringInSlice(component, components) {
			continue
		}
		newComponents = append(newComponents, component)
	}

	m.Components = newComponents
	return mf.Write(m)
}
