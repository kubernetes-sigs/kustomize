// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
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
	c.Flags().BoolVar(&r.Markdown, "markdown", false,
		"output as github markdown")
	c.Flags().BoolVar(&r.IncludeSubst, "include-subst", false,
		"include substitutions in the output")
	fixDocs(parent, c)
	r.Command = c
	return r
}

func ListSettersCommand(parent string) *cobra.Command {
	return NewListSettersRunner(parent).Command
}

type ListSettersRunner struct {
	Command      *cobra.Command
	Lookup       setters.LookupSettersDeprecated
	List         setters2.List
	Markdown     bool
	IncludeSubst bool
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
		if err := r.ListSetters(c, args); err != nil {
			return err
		}
		if r.IncludeSubst {
			return r.ListSubstitutions(c, args)
		}
		return nil
	}

	return handleError(c, lookup(r.Lookup, c, args))
}

func (r *ListSettersRunner) ListSetters(c *cobra.Command, args []string) error {
	// use setters v2
	path, err := ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}
	if err := r.List.ListSetters(path, args[0]); err != nil {
		return err
	}
	table := newTable(c.OutOrStdout(), r.Markdown)
	table.SetHeader([]string{"NAME", "VALUE", "SET BY", "DESCRIPTION", "COUNT", "REQUIRED"})
	for i := range r.List.Setters {
		s := r.List.Setters[i]
		v := s.Value

		// if the setter is for a list, populate the values
		if len(s.ListValues) > 0 {
			v = strings.Join(s.ListValues, ",")
			v = fmt.Sprintf("[%s]", v)
		}
		var required string
		if s.Required {
			required = "Yes"
		} else {
			required = "No"
		}
		table.Append([]string{
			s.Name, v, s.SetBy, s.Description, fmt.Sprintf("%d", s.Count), required})
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

func (r *ListSettersRunner) ListSubstitutions(c *cobra.Command, args []string) error {
	// use setters v2
	path, err := ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}
	if err := r.List.ListSubst(path); err != nil {
		return err
	}
	table := newTable(c.OutOrStdout(), r.Markdown)
	b := tablewriter.Border{Top: true}
	table.SetBorders(b)

	table.SetHeader([]string{"SUBSTITUTION", "PATTERN", "REFERENCES"})
	for i := range r.List.Substitutions {
		s := r.List.Substitutions[i]
		refs := ""
		for _, value := range s.Values {
			// trim setter and substitution prefixes
			ref := strings.TrimPrefix(
				strings.TrimPrefix(value.Ref, fieldmeta.DefinitionsPrefix+fieldmeta.SetterDefinitionPrefix),
				fieldmeta.DefinitionsPrefix+fieldmeta.SubstitutionDefinitionPrefix)
			refs = refs + "," + ref
		}
		refs = fmt.Sprintf("[%s]", strings.TrimPrefix(refs, ","))
		table.Append([]string{
			s.Name, s.Pattern, refs})
	}
	if len(r.List.Substitutions) == 0 {
		return nil
	}
	table.Render()

	return nil
}

func newTable(o io.Writer, m bool) *tablewriter.Table {
	table := tablewriter.NewWriter(o)
	table.SetRowLine(false)
	if m {
		// markdown format
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	} else {
		table.SetBorder(false)
		table.SetHeaderLine(false)
		table.SetColumnSeparator(" ")
		table.SetCenterSeparator(" ")
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	return table
}
