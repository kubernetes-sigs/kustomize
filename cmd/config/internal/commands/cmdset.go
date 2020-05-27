// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/setters"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewSetRunner returns a command runner.
func NewSetRunner(parent string) *SetRunner {
	r := &SetRunner{}
	c := &cobra.Command{
		Use:     "set DIR NAME [VALUE]",
		Args:    cobra.MinimumNArgs(1),
		Short:   commands.SetShort,
		Long:    commands.SetLong,
		Example: commands.SetExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c
	c.Flags().StringVar(&r.Set.Name, "name", "", "name of the setter")
	c.Flags().StringVar(&r.Set.Value, "value", "", "value of the setter to set")
	c.Flags().StringArrayVar(&r.Set.ListValues, "list-value", []string{}, "List value to be set")
	c.Flags().StringVar(&r.Perform.SetBy, "set-by", "",
		"annotate the field with who set it")
	c.Flags().StringVar(&r.Perform.Description, "description", "",
		"annotate the field with a description of its value")
	c.Flags().StringVar(&setterVersion, "version", "",
		"use this version of the setter format")
	c.Flags().MarkHidden("version")

	return r
}

var setterVersion string

func SetCommand(parent string) *cobra.Command {
	return NewSetRunner(parent).Command
}

type SetRunner struct {
	Command     *cobra.Command
	Lookup      setters.LookupSetters
	Perform     setters.PerformSetters
	Set         settersutil.FieldSetter
	OpenAPIFile string
}

func initSetterVersion(c *cobra.Command, args []string) error {
	setterVersion = "v2"
	l := setters.LookupSetters{}

	// backwards compatibility for resources with setter v1
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.LocalPackageReader{PackagePath: args[0]}},
		Filters: []kio.Filter{&l},
	}.Execute()
	if err != nil {
		return err
	}
	if len(l.SetterCounts) > 0 {
		setterVersion = "v1"
	}

	return nil
}

func (r *SetRunner) preRunE(c *cobra.Command, args []string) error {
	r.Perform.Name = r.Set.Name
	r.Lookup.Name = r.Set.Name
	r.Perform.Value = r.Set.Value

	if setterVersion == "" {
		if len(r.Set.ListValues) == 0 {
			setterVersion = "v1"
		} else if err := initSetterVersion(c, args); err != nil {
			return err
		}
	}
	if setterVersion == "v2" {
		var err error
		r.Set.Description = r.Perform.Description
		r.Set.SetBy = r.Perform.SetBy
		r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SetRunner) runE(c *cobra.Command, args []string) error {
	if setterVersion == "v2" {
		count, err := r.Set.Set(r.OpenAPIFile, args[0])
		fmt.Fprintf(c.OutOrStdout(), "set %d fields\n", count)
		return handleError(c, err)
	}
	if len(args) == 3 {
		return handleError(c, r.perform(c, args))
	}
	return handleError(c, lookup(r.Lookup, c, args))
}

func lookup(l setters.LookupSetters, c *cobra.Command, args []string) error {
	// lookup the setters
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.LocalPackageReader{PackagePath: args[0]}},
		Filters: []kio.Filter{&l},
	}.Execute()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(c.OutOrStdout())
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator(" ")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{
		"NAME", "DESCRIPTION", "VALUE", "TYPE", "COUNT", "SETBY",
	})
	for i := range l.SetterCounts {
		s := l.SetterCounts[i]
		v := s.Value
		if s.Value == "" {
			v = s.Value
		}
		table.Append([]string{
			s.Name,
			"'" + s.Description + "'",
			v,
			fmt.Sprintf("%v", s.Type),
			fmt.Sprintf("%d", s.Count),
			s.SetBy,
		})
	}
	table.Render()

	if len(l.SetterCounts) == 0 {
		// exit non-0 if no matching setters are found
		os.Exit(1)
	}
	return nil
}

// perform the setters
func (r *SetRunner) perform(c *cobra.Command, args []string) error {
	rw := &kio.LocalPackageReadWriter{
		PackagePath: args[0],
	}
	// perform the setters in the package
	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&r.Perform},
		Outputs: []kio.Writer{rw},
	}.Execute()
	if err != nil {
		return err
	}
	fmt.Fprintf(c.OutOrStdout(), "set %d fields\n", r.Perform.Count)
	return nil
}
