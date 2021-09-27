// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeResourceOptions struct {
	resourceFilePaths []string
}

// newCmdRemoveResource remove the name of a file containing a resource to the kustomization file.
func newCmdRemoveResource(fSys filesys.FileSystem) *cobra.Command {
	var o removeResourceOptions

	cmd := &cobra.Command{
		Use: "resource",
		Short: "Removes one or more resource file paths from " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		remove resource my-resource.yml
		remove resource resource1.yml resource2.yml resource3.yml
		remove resource resources/*.yml
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunRemoveResource(fSys)
		},
	}
	return cmd
}

// Validate validates removeResource command.
func (o *removeResourceOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a resource file")
	}
	o.resourceFilePaths = args
	return nil
}

// RunRemoveResource runs Resource command (do real work).
func (o *removeResourceOptions) RunRemoveResource(fSys filesys.FileSystem) error {

	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	resources, err := globPatterns(m.Resources, o.resourceFilePaths)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		return nil
	}

	newResources := make([]string, 0, len(m.Resources))
	for _, resource := range m.Resources {
		if kustfile.StringInSlice(resource, resources) {
			continue
		}
		newResources = append(newResources, resource)
	}

	m.Resources = newResources
	return mf.Write(m)
}

func globPatterns(resources []string, patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		for _, resource := range resources {
			match, err := filepath.Match(pattern, resource)
			if err != nil {
				return nil, err
			}
			if !match {
				continue
			}
			result = append(result, resource)
		}
	}
	return result, nil
}
