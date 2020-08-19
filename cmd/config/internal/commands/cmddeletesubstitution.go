// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewDeleteRunner returns a command runner.
func NewDeleteSubstitutionRunner(parent string) *DeleteSubstitutionRunner {
	r := &DeleteSubstitutionRunner{}
	c := &cobra.Command{
		Use:     "delete-subst DIR NAME",
		Args:    cobra.ExactArgs(2),
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c

	return r
}

func DeleteSubstitutionCommand(parent string) *cobra.Command {
	return NewDeleteSubstitutionRunner(parent).Command
}

type DeleteSubstitutionRunner struct {
	Command            *cobra.Command
	DeleteSubstitution settersutil.DeleterCreator
	OpenAPIFile        string
}

func (r *DeleteSubstitutionRunner) preRunE(c *cobra.Command, args []string) error {
	var err error
	r.DeleteSubstitution.Name = args[1]
	r.DeleteSubstitution.DefinitionPrefix = fieldmeta.SubstitutionDefinitionPrefix

	r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}

	if err := openapi.AddSchemaFromFile(r.OpenAPIFile); err != nil {
		return err
	}

	return nil
}

func (r *DeleteSubstitutionRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, r.delete(c, args))
}

func (r *DeleteSubstitutionRunner) delete(c *cobra.Command, args []string) error {
	return r.DeleteSubstitution.Delete(r.OpenAPIFile, args[0])
}
