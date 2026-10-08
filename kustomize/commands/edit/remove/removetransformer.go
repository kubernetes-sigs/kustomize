// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeTransformerOptions struct {
	transformerFilePaths []string
}

// newCmdRemoveTransformer remove the name of a file containing a transformer to the kustomization file.
func newCmdRemoveTransformer(fSys filesys.FileSystem) *cobra.Command {
	var o removeTransformerOptions

	cmd := &cobra.Command{
		Use: "transformer",
		Short: "Removes one or more transformers from " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		remove transformer my-transformer.yml
		remove transformer transformer1.yml transformer2.yml transformer3.yml
		remove transformer transformers/*.yml
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunRemoveTransformer(fSys)
		},
	}
	return cmd
}

// Validate validates removeTransformer command.
func (o *removeTransformerOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a transformer file")
	}
	o.transformerFilePaths = args
	return nil
}

// RunRemoveTransformer runs Transformer command (do real work).
func (o *removeTransformerOptions) RunRemoveTransformer(fSys filesys.FileSystem) error {
	return removePathsFromKustomization(fSys, o.transformerFilePaths, func(m *types.Kustomization) *[]string {
		return &m.Transformers
	})
}
