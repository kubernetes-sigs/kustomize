// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addComponentOptions struct {
	componentFilePaths []string
}

// newCmdAddComponent adds the name of a file containing a component to the kustomization file.
func newCmdAddComponent(fSys filesys.FileSystem) *cobra.Command {
	var o addComponentOptions

	cmd := &cobra.Command{
		Use:   "component",
		Short: "Add the name of a file containing a component to the kustomization file.",
		Example: `
		add component {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunAddComponent(fSys)
		},
	}
	return cmd
}

// Validate validates addComponent command.
func (o *addComponentOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a component file")
	}
	o.componentFilePaths = args
	return nil
}

// RunAddComponent runs addComponent command (do real work).
func (o *addComponentOptions) RunAddComponent(fSys filesys.FileSystem) error {
	components, err := util.GlobPatternsWithLoader(fSys, loader.NewFileLoaderAtCwd(fSys), o.componentFilePaths)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return nil
	}

	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	for _, component := range components {
		if mf.GetPath() != component {
			if kustfile.StringInSlice(component, m.Components) {
				log.Printf("component %s already in kustomization file", component)
				continue
			}
			m.Components = append(m.Components, component)
		}
	}

	return mf.Write(m)
}
