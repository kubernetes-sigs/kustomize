// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
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
		Deprecated: "setter commands will no longer be available in kustomize v5.\n" +
			"See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.",
	}
	c.Flags().BoolVar(&r.Markdown, "markdown", false,
		"output as github markdown")
	c.Flags().BoolVar(&r.IncludeSubst, "include-subst", false,
		"include substitutions in the output")
	c.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"list setters recursively in all the nested subpackages")
	runner.FixDocs(parent, c)
	r.Command = c
	return r
}

func ListSettersCommand(parent string) *cobra.Command {
	return NewListSettersRunner(parent).Command
}

type ListSettersRunner struct {
	Command            *cobra.Command
	List               setters2.List
	Markdown           bool
	IncludeSubst       bool
	RecurseSubPackages bool
	Name               string
}

func (r *ListSettersRunner) preRunE(c *cobra.Command, args []string) error {
	if len(args) > 1 {
		r.Name = args[1]
	}

	return nil
}

func (r *ListSettersRunner) runE(c *cobra.Command, args []string) error {
	e := runner.ExecuteCmdOnPkgs{
		NeedOpenAPI:        true,
		Writer:             c.OutOrStdout(),
		RootPkgPath:        args[0],
		RecurseSubPackages: r.RecurseSubPackages,
		CmdRunner:          r,
	}

	err := e.Execute()
	if err != nil {
		return runner.HandleError(c, err)
	}
	return nil
}

func (r *ListSettersRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	sc, err := openapi.SchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName()))
	if err != nil {
		return err
	}
	r.List = setters2.List{
		Name:            r.Name,
		OpenAPIFileName: ext.KRMFileName(),
		SettersSchema:   sc,
	}
	openAPIPath := filepath.Join(pkgPath, ext.KRMFileName())
	if err := r.ListSetters(w, openAPIPath, pkgPath); err != nil {
		return err
	}
	if r.IncludeSubst {
		if err := r.ListSubstitutions(w, openAPIPath); err != nil {
			return err
		}
	}
	return nil
}

func (r *ListSettersRunner) ListSetters(w io.Writer, openAPIPath, resourcePath string) error {
	// use setters v2
	if err := r.List.ListSetters(openAPIPath, resourcePath); err != nil {
		return err
	}
	table := newTable(w, r.Markdown)
	table.SetHeader([]string{"NAME", "VALUE", "SET BY", "DESCRIPTION", "COUNT", "REQUIRED", "IS SET"})
	for i := range r.List.Setters {
		s := r.List.Setters[i]
		v := s.Value

		// if the setter is for a list, populate the values
		if len(s.ListValues) > 0 {
			v = strings.Join(s.ListValues, ",")
			v = fmt.Sprintf("[%s]", v)
		}
		required := "No"
		if s.Required {
			required = "Yes"
		}
		isSet := "No"
		if s.IsSet {
			isSet = "Yes"
		}

		table.Append([]string{
			s.Name, v, s.SetBy, s.Description, fmt.Sprintf("%d", s.Count), required, isSet})
	}
	table.Render()

	if len(r.List.Setters) == 0 {
		// exit non-0 if no matching setters are found
		if runner.ExitOnError {
			os.Exit(1)
		}
	}
	return nil
}

func (r *ListSettersRunner) ListSubstitutions(w io.Writer, openAPIPath string) error {
	// use setters v2
	if err := r.List.ListSubst(openAPIPath); err != nil {
		return err
	}
	table := newTable(w, r.Markdown)
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
