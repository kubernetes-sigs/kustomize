// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/util"
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
			err = o.Complete(cmd, args)
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

// Complete completes addResource command.
func (o *addResourceOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddResource runs addResource command (do real work).
func (o *addResourceOptions) RunAddResource(fSys filesys.FileSystem) error {
	resources, err := util.GlobPatterns(fSys, o.resourceFilePaths)
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
		if kustfile.StringInSlice(resource, m.Resources) {
			log.Printf("resource %s already in kustomization file", resource)
			continue
		}
		m.Resources = append(m.Resources, resource)
	}

	return mf.Write(m)
}
