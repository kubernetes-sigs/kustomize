// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"errors"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type publishOptions struct {
	registry string
	noVerify bool
}

// NewCmdEdit returns an instance of 'edit' subcommand.
func NewCmdPublish(
	fSys filesys.FileSystem, v ifc.Validator, rf *resource.Factory,
) *cobra.Command {
	var o publishOptions

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publishes a kustomization resource to an OCI registry",
		Long:  "",
		Example: `
		publish <registry>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunPublish(fSys)
		},

		Args: cobra.MinimumNArgs(1),
	}
	cmd.Flags().BoolVar(&o.noVerify, "no-verify", false,
		"skip validation for resources",
	)
	return cmd
}

// Validate validates addResource command.
func (o *publishOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a registry")
	}
	o.registry = args[0]
	return nil
}

// RunAddResource runs addResource command (do real work).
func (o *publishOptions) RunPublish(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}

	for _, r := range m.Resources {
		println(r)
	}

	return nil
}
