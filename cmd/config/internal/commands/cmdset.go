// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/setters"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewSetRunner returns a command runner.
func NewSetRunner(parent string) *SetRunner {
	r := &SetRunner{}
	c := &cobra.Command{
		Use:     "set DIR NAME --values [VALUE]",
		Args:    cobra.MinimumNArgs(2),
		Short:   commands.SetShort,
		Long:    commands.SetLong,
		Example: commands.SetExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	runner.FixDocs(parent, c)
	r.Command = c
	c.Flags().StringArrayVar(&r.Values, "values", []string{},
		"optional flag, the values of the setter to be set to")
	c.Flags().StringVar(&r.Perform.SetBy, "set-by", "",
		"annotate the field with who set it")
	c.Flags().StringVar(&r.Perform.Description, "description", "",
		"annotate the field with a description of its value")
	c.Flags().StringVar(&setterVersion, "version", "",
		"use this version of the setter format")
	c.Flags().BoolVarP(&r.Set.RecurseSubPackages, "recurse-subpackages", "R", false,
		"sets recursively in all the nested subpackages")
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
	Values      []string
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
	valueFlagSet := c.Flag("values").Changed

	if valueFlagSet && len(args) > 2 {
		return errors.Errorf("value should set either from flag or arg")
	}

	if len(args) > 1 {
		r.Perform.Name = args[1]
		r.Lookup.Name = args[1]
	}

	if valueFlagSet {
		r.Perform.Value = r.Values[0]
	} else if len(args) > 2 {
		r.Perform.Value = args[2]
	}

	if setterVersion == "" {
		if len(args) < 2 || len(args) < 3 && !valueFlagSet {
			setterVersion = "v1"
		} else if err := initSetterVersion(c, args); err != nil {
			return err
		}
	}
	if setterVersion == "v2" {
		r.Set.Name = args[1]
		if valueFlagSet {
			r.Set.Value = r.Values[0]
		} else {
			r.Set.Value = args[2]
		}

		// set remaining values as list values
		if valueFlagSet && len(r.Values) > 1 {
			r.Set.ListValues = r.Values[1:]
		} else if !valueFlagSet && len(args) > 3 {
			r.Set.ListValues = args[3:]
		}

		r.Set.Description = r.Perform.Description
		r.Set.SetBy = r.Perform.SetBy
		r.OpenAPIFile = filepath.Join(args[0], ext.KRMFileName())
	}
	return nil
}

func (r *SetRunner) runE(c *cobra.Command, args []string) error {
	if setterVersion == "v2" {
		e := runner.ExecuteCmdOnPkgs{
			NeedOpenAPI:        true,
			Writer:             c.OutOrStdout(),
			RootPkgPath:        args[0],
			RecurseSubPackages: r.Set.RecurseSubPackages,
			CmdRunner:          r,
		}
		err := e.Execute()
		if err != nil {
			return runner.HandleError(c, err)
		}
		return nil
	}
	if len(args) > 2 || c.Flag("values").Changed {
		return runner.HandleError(c, r.perform(c, args))
	}
	return runner.HandleError(c, lookup(r.Lookup, c, args))
}

func (r *SetRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	r.Set = settersutil.FieldSetter{
		Name:               r.Set.Name,
		Value:              r.Set.Value,
		ListValues:         r.Set.ListValues,
		Description:        r.Set.Description,
		SetBy:              r.Set.SetBy,
		Count:              0,
		OpenAPIPath:        filepath.Join(pkgPath, ext.KRMFileName()),
		OpenAPIFileName:    ext.KRMFileName(),
		ResourcesPath:      pkgPath,
		RecurseSubPackages: r.Set.RecurseSubPackages,
		IsSet:              true,
	}
	count, err := r.Set.Set()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.Set.RecurseSubPackages {
			return err
		} else {
			// print error message and continue if RecurseSubPackages is true
			fmt.Fprintf(w, "%s\n", err.Error())
		}
	} else {
		fmt.Fprintf(w, "set %d field(s) of setter %q to value %q\n", count, r.Set.Name, r.Set.Value)
	}
	return nil
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
