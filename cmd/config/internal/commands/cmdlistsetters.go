// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/setters"
	"sigs.k8s.io/kustomize/kyaml/setters2"
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
	List    setters2.List
}

func (r *ListSettersRunner) preRunE(c *cobra.Command, args []string) error {
	if len(args) > 1 {
		r.Lookup.Name = args[1]
		r.List.Name = args[1]
	}

	initSetterVersion(c, args)
	return nil
}

func (r *ListSettersRunner) runE(c *cobra.Command, args []string) error {
	if setterVersion == "v2" {
		// use setters v2
		path, err := ext.GetOpenAPIFile(args)
		if err != nil {
			return err
		}
		if err := r.List.List(path, args[0]); err != nil {
			return err
		}
		table := newTable(c.OutOrStdout())
		table.SetHeader([]string{"NAME", "VALUE", "SET BY", "DESCRIPTION", "COUNT"})
		for i := range r.List.Setters {
			s := r.List.Setters[i]
			table.Append([]string{
				s.Name, s.Value, s.SetBy, s.Description, fmt.Sprintf("%d", s.Count)})
		}
		table.Render()

		if len(r.List.Setters) == 0 {
			// exit non-0 if no matching setters are found
			if ExitOnError {
				os.Exit(1)
			}
		}
		return nil
	}

	return handleError(c, lookup(r.Lookup, c, args))
}

func newTable(o io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(o)
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator(" ")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	return table
}
