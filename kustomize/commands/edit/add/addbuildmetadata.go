// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addBuildMetadataOptions struct {
	buildMetadataOptions []string
}

// newCmdAddBuildMetadata adds options to the kustomization's buildMetada field.
func newCmdAddBuildMetadata(fSys filesys.FileSystem) *cobra.Command {
	var o addBuildMetadataOptions

	cmd := &cobra.Command{
		Use:   "buildmetadata",
		Short: "Adds one or more buildMetadata options to the kustomization.yaml in the current directory",
		Long:  `Adds one or more buildMetadata options to the kustomization.yaml in the current directory.
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
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunAddBuildMetadata(fSys)
		},
	}
	return cmd
}

// Validate validates addBuildMetadata command.
func (o *addBuildMetadataOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a buildMetadata option")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: %s; to provide multiple buildMetadata options, please separate options by comma", args)
	}
	o.buildMetadataOptions = strings.Split(args[0], ",")
	for _, opt := range o.buildMetadataOptions {
		if !kustfile.StringInSlice(opt, types.BuildMetadataOptions) {
			return fmt.Errorf("invalid buildMetadata option: %s", opt)
		}
	}
	return nil
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
