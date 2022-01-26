// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeBuildMetadataOptions struct {
	*util.BuildMetadataValidator
	buildMetadataOptions []string
}

// newCmdRemoveBuildMetadata removes options to the kustomization's buildMetada field.
func newCmdRemoveBuildMetadata(fSys filesys.FileSystem) *cobra.Command {
	var o removeBuildMetadataOptions

	cmd := &cobra.Command{
		Use:   "buildmetadata",
		Short: "Removes one or more buildMetadata options to the kustomization.yaml in the current directory",
		Long: `Removes one or more buildMetadata options to the kustomization.yaml in the current directory.
The following options are valid:
  - originAnnotations
  - transformerAnnotations
  - managedByLabel
originAnnotations will remove the annotation config.kubernetes.io/origin to each resource, describing where 
each resource originated from.
transformerAnnotations will remove the annotation alpha.config.kubernetes.io/transformations to each resource,
describing the transformers that have acted upon the resource.
managedByLabel will remove the label app.kubernetes.io/managed-by to each resource, describing which version
of kustomize managed the resource.`,
		Example: `
		remove buildmetadata {option1},{option2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			o.buildMetadataOptions, err = o.BuildMetadataValidator.Validate(args)
			if err != nil {
				return err
			}
			return o.RunRemoveBuildMetadata(fSys)
		},
	}
	return cmd
}

// RunRemoveBuildMetadata runs removeBuildMetadata command (do real work).
func (o *removeBuildMetadataOptions) RunRemoveBuildMetadata(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	var newOptions []string
	for _, opt := range m.BuildMetadata {
		if !kustfile.StringInSlice(opt, o.buildMetadataOptions) {
			newOptions = append(newOptions, opt)
		}
	}
	m.BuildMetadata = newOptions
	if len(m.BuildMetadata) == 0 {
		m.BuildMetadata = nil
	}
	return mf.Write(m)
}
