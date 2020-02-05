// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/setters"
)

// NewCreateSetterRunner returns a command runner.
func NewCreateSetterRunner(parent string) *CreateSetterRunner {
	r := &CreateSetterRunner{}
	set := &cobra.Command{
		Use:     "create-setter DIR NAME VALUE",
		Args:    cobra.ExactArgs(3),
		Short:   commands.CreateSetterShort,
		Long:    commands.CreateSetterLong,
		Example: commands.CreateSetterExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	set.Flags().StringVar(&r.Set.SetPartialField.SetBy, "set-by", "",
		"set the setBy annotation.")
	set.Flags().StringVar(&r.Set.SetPartialField.Description, "description", "",
		"set the description of the field value.")
	set.Flags().StringVar(&r.Set.SetPartialField.Field, "field", "",
		"name of the field to set -- e.g. --field port")
	set.Flags().StringVar(&r.Set.ResourceMeta.Name, "name", "",
		"name of the Resource on which to create the setter.")
	set.Flags().StringVar(&r.Set.ResourceMeta.Kind, "kind", "",
		"kind of the Resource on which to create the setter.")
	set.Flags().StringVar(&r.Set.SetPartialField.Type, "type", "",
		"valid OpenAPI field type -- e.g. integer,boolean,string.")
	set.Flags().BoolVar(&r.Set.SetPartialField.Partial, "partial", false,
		"create a partial setter for only part of the field value.")
	fixDocs(parent, set)
	set.MarkFlagRequired("type")
	r.Command = set
	return r
}

func CreateSetterCommand(parent string) *cobra.Command {
	return NewCreateSetterRunner(parent).Command
}

type CreateSetterRunner struct {
	Command *cobra.Command
	Set     setters.CreateSetter
}

func (r *CreateSetterRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, r.set(c, args))
}

func (r *CreateSetterRunner) preRunE(c *cobra.Command, args []string) error {
	r.Set.SetPartialField.Setter.Name = args[1]
	r.Set.SetPartialField.Setter.Value = args[2]
	return nil
}

func (r *CreateSetterRunner) set(c *cobra.Command, args []string) error {
	rw := &kio.LocalPackageReadWriter{PackagePath: args[0]}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&r.Set},
		Outputs: []kio.Writer{rw}}.Execute()
	if err != nil {
		return err
	}
	return nil
}
