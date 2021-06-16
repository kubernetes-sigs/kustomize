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

type addResourceOptions struct {
	resourceFilePaths []string
}

// newCmdAddResource adds the name of a file containing a resource to the kustomization file.
func newCmdAddResource(fSys filesys.FileSystem) *cobra.Command {
	var o addResourceOptions

	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Add the name of a file containing a resource to the kustomization file.",
		Example: `
		add resource {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunAddResource(fSys)
		},
	}
	return cmd
}

// Validate validates addResource command.
func (o *addResourceOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a resource file")
	}
	o.resourceFilePaths = args
	return nil
}

// RunAddResource runs addResource command (do real work).
func (o *addResourceOptions) RunAddResource(fSys filesys.FileSystem) error {
	resources, err := util.GlobPatternsWithLoader(fSys, loader.NewFileLoaderAtCwd(fSys), o.resourceFilePaths)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
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

	for _, resource := range resources {
		if mf.GetPath() != resource {
			if kustfile.StringInSlice(resource, m.Resources) {
				log.Printf("resource %s already in kustomization file", resource)
				continue
			}
			m.Resources = append(m.Resources, resource)
		}
	}

	return mf.Write(m)
}
