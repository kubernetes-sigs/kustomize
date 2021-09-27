// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setNamespaceOptions struct {
	namespace string
	validator ifc.Validator
}

// newCmdSetNamespace sets the value of the namespace field in the kustomization.
func newCmdSetNamespace(fSys filesys.FileSystem, v ifc.Validator) *cobra.Command {
	var o setNamespaceOptions

	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "Sets the value of the namespace field in the kustomization file",
		Example: `
The command
	set namespace staging
will add the field "namespace: staging" to the kustomization file if it doesn't exist,
and overwrite the value with "staging" if the field does exist.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.validator = v
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunSetNamespace(fSys)
		},
	}
	return cmd
}

// Validate validates setNamespace command.
func (o *setNamespaceOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify exactly one namespace value")
	}
	ns := args[0]
	if errs := o.validator.ValidateNamespace(ns); len(errs) != 0 {
		return fmt.Errorf("%q is not a valid namespace name: %s", ns, strings.Join(errs, ";"))
	}
	o.namespace = ns
	return nil
}

// RunSetNamespace runs setNamespace command (does real work).
func (o *setNamespaceOptions) RunSetNamespace(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	m.Namespace = o.namespace
	return mf.Write(m)
}
