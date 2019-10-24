// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"errors"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
)

type setNamePrefixOptions struct {
	prefix string
}

// newCmdSetNamePrefix sets the value of the namePrefix field in the kustomization.
func newCmdSetNamePrefix(fSys filesys.FileSystem) *cobra.Command {
	var o setNamePrefixOptions

	cmd := &cobra.Command{
		Use:   "nameprefix",
		Short: "Sets the value of the namePrefix field in the kustomization file.",
		Example: `
The command
  set nameprefix acme-
will add the field "namePrefix: acme-" to the kustomization file if it doesn't exist,
and overwrite the value with "acme-" if the field does exist.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunSetNamePrefix(fSys)
		},
	}
	return cmd
}

// Validate validates setNamePrefix command.
func (o *setNamePrefixOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify exactly one prefix value")
	}
	// TODO: add further validation on the value.
	o.prefix = args[0]
	return nil
}

// Complete completes setNamePrefix command.
func (o *setNamePrefixOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunSetNamePrefix runs setNamePrefix command (does real work).
func (o *setNamePrefixOptions) RunSetNamePrefix(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	m.NamePrefix = o.prefix
	return mf.Write(m)
}
