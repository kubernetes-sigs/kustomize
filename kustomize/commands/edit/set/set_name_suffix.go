// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"errors"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setNameSuffixOptions struct {
	suffix string
}

// newCmdSetNameSuffix sets the value of the nameSuffix field in the kustomization.
func newCmdSetNameSuffix(fSys filesys.FileSystem) *cobra.Command {
	var o setNameSuffixOptions

	cmd := &cobra.Command{
		Use:   "namesuffix",
		Short: "Sets the value of the nameSuffix field in the kustomization file.",
		Example: `
The command
  set namesuffix -- -acme
will add the field "nameSuffix: -acme" to the kustomization file if it doesn't exist,
and overwrite the value with "-acme" if the field does exist.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunSetNameSuffix(fSys)
		},
	}
	return cmd
}

// Validate validates setNameSuffix command.
func (o *setNameSuffixOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify exactly one suffix value")
	}
	// TODO: add further validation on the value.
	o.suffix = args[0]
	return nil
}

// RunSetNameSuffix runs setNameSuffix command (does real work).
func (o *setNameSuffixOptions) RunSetNameSuffix(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	m.NameSuffix = o.suffix
	return mf.Write(m)
}
