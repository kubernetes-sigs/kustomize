// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addBuildMetadataOptions struct {
	*util.BuildMetadataValidator
	buildMetadataOptions []string
}

// newCmdAddBuildMetadata adds options to the kustomization's buildMetada field.
func newCmdAddBuildMetadata(fSys filesys.FileSystem) *cobra.Command {
	var o addBuildMetadataOptions

	cmd := &cobra.Command{
		Use:   "buildmetadata",
		Short: "Adds one or more buildMetadata options to the kustomization.yaml in the current directory",
		Long: `Adds one or more buildMetadata options to the kustomization.yaml in the current directory.
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
		add buildmetadata {option1},{option2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			o.buildMetadataOptions, err = o.BuildMetadataValidator.Validate(args)
			if err != nil {
				return err
			}
			return o.RunAddBuildMetadata(fSys)
		},
	}
	return cmd
}

// RunAddBuildMetadata runs addBuildMetadata command (do real work).
func (o *addBuildMetadataOptions) RunAddBuildMetadata(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	for _, opt := range o.buildMetadataOptions {
		if kustfile.StringInSlice(opt, m.BuildMetadata) {
			return fmt.Errorf("buildMetadata option %s already in kustomization file", opt)
		}
		m.BuildMetadata = append(m.BuildMetadata, opt)
	}
	return mf.Write(m)
}
