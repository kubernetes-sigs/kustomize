// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/setters"
)

// NewListSettersRunner returns a command runner.
func NewListSettersRunner(parent string) *ListSettersRunner {
	r := &ListSettersRunner{}
	c := &cobra.Command{
		Use:     "list-setters DIR [NAME]",
		Args:    cobra.RangeArgs(1, 2),
		Short:   commands.ListSettersShort,
		Long:    commands.ListSettersLong,
		Example: commands.ListSettersExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c
	return r
}

func ListSettersCommand(parent string) *cobra.Command {
	return NewListSettersRunner(parent).Command
}

type ListSettersRunner struct {
	Command *cobra.Command
	Lookup  setters.LookupSetters
}

func (r *ListSettersRunner) preRunE(c *cobra.Command, args []string) error {
	if len(args) > 1 {
		r.Lookup.Name = args[1]
	}
	return nil
}

func (r *ListSettersRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, lookup(r.Lookup, c, args))
}
