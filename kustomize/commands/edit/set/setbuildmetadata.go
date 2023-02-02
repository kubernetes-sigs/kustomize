// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setBuildMetadataOptions struct {
	*util.BuildMetadataValidator
	buildMetadataOptions []string
}

// newCmdSetBuildMetadata sets options in the kustomization's buildMetada field.
func newCmdSetBuildMetadata(fSys filesys.FileSystem) *cobra.Command {
	var o setBuildMetadataOptions

	cmd := &cobra.Command{
		Use:   "buildmetadata",
		Short: "Sets one or more buildMetadata options to the kustomization.yaml in the current directory",
		Long: `Sets one or more buildMetadata options to the kustomization.yaml in the current directory.
Existing options in the buildMetadata field will be replaced entirely by the new options set by this command.
The following options are valid:
  - originAnnotations
  - transformerAnnotations
  - managedByLabel
originAnnotations will add the annotation config.kubernetes.io/origin to each resource, describing where 
each resource originated from.
transformerAnnotations will add the annotation alpha.config.kubernetes.io/transformations to each resource,
describing the transformers that have acted upon the resource.
managedByLabel will add the label app.kubernetes.io/managed-by to each resource, describing which version
of kustomize managed the resource.`,
		Example: `
		set buildmetadata {option1},{option2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			o.buildMetadataOptions, err = o.BuildMetadataValidator.Validate(args)
			if err != nil {
				return err
			}
			return o.RunSetBuildMetadata(fSys)
		},
	}
	return cmd
}

// RunSetBuildMetadata runs setBuildMetadata command (do real work).
func (o *setBuildMetadataOptions) RunSetBuildMetadata(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	m.BuildMetadata = o.buildMetadataOptions
	return mf.Write(m)
}
