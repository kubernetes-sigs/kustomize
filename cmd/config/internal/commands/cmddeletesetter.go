// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewDeleteRunner returns a command runner.
func NewDeleteSetterRunner(parent string) *DeleteSetterRunner {
	r := &DeleteSetterRunner{}
	c := &cobra.Command{
		Use:     "delete-setter DIR NAME",
		Args:    cobra.MinimumNArgs(2),
		Short:   commands.SetShort,
		Long:    commands.SetLong,
		Example: commands.SetExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c

	return r
}

func DeleteSetterCommand(parent string) *cobra.Command {
	return NewDeleteSetterRunner(parent).Command
}

type DeleteSetterRunner struct {
	Command      *cobra.Command
	DeleteSetter settersutil.DeleterCreator
	OpenAPIFile  string
}

func (r *DeleteSetterRunner) preRunE(c *cobra.Command, args []string) error {
	var err error
	r.DeleteSetter.Name = args[1]

	r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}

	if err := openapi.AddSchemaFromFile(r.OpenAPIFile); err != nil {
		return err
	}

	return nil
}

func (r *DeleteSetterRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, r.delete(c, args))
}

func (r *DeleteSetterRunner) delete(c *cobra.Command, args []string) error {
	return r.DeleteSetter.Delete(r.OpenAPIFile, args[0])
}
