// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/errors"
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
		Deprecated: "setter commands will no longer be available in kustomize v5.\n" +
			"See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.",
	}
	runner.FixDocs(parent, c)
	r.Command = c
	c.Flags().StringArrayVar(&r.Values, "values", []string{},
		"optional flag, the values of the setter to be set to")
	c.Flags().StringVar(&r.SetBy, "set-by", "",
		"annotate the field with who set it")
	c.Flags().StringVar(&r.Description, "description", "",
		"annotate the field with a description of its value")
	c.Flags().StringVar(&setterVersion, "version", "",
		"use this version of the setter format")
	c.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"sets recursively in all the nested subpackages")
	c.Flags().MarkHidden("version")

	return r
}

var setterVersion string

func SetCommand(parent string) *cobra.Command {
	return NewSetRunner(parent).Command
}

type SetRunner struct {
	Command            *cobra.Command
	Set                settersutil.FieldSetter
	OpenAPIFile        string
	Values             []string
	SetBy              string
	Description        string
	Name               string
	Value              string
	ListValues         []string
	RecurseSubPackages bool
}

func (r *SetRunner) preRunE(c *cobra.Command, args []string) error {
	valueFlagSet := c.Flag("values").Changed

	if valueFlagSet && len(args) > 2 {
		return errors.Errorf("value should set either from flag or arg")
	}

	// make sure that the value is provided either through values flag or as an arg
	if !valueFlagSet && len(args) < 3 {
		return errors.Errorf("value must be provided either from flag or arg")
	}

	r.Name = args[1]
	if valueFlagSet {
		r.Value = r.Values[0]
	} else {
		r.Value = args[2]
	}

	// set remaining values as list values
	if valueFlagSet && len(r.Values) > 1 {
		r.ListValues = r.Values[1:]
	} else if !valueFlagSet && len(args) > 3 {
		r.ListValues = args[3:]
	}

	r.OpenAPIFile = filepath.Join(args[0], ext.KRMFileName())
	return nil
}

func (r *SetRunner) runE(c *cobra.Command, args []string) error {
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

func (r *SetRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	r.Set = settersutil.FieldSetter{
		Name:               r.Name,
		Value:              r.Value,
		ListValues:         r.ListValues,
		Description:        r.Description,
		SetBy:              r.SetBy,
		Count:              0,
		OpenAPIPath:        filepath.Join(pkgPath, ext.KRMFileName()),
		OpenAPIFileName:    ext.KRMFileName(),
		ResourcesPath:      pkgPath,
		RecurseSubPackages: r.RecurseSubPackages,
		IsSet:              true,
	}
	count, err := r.Set.Set()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.Set.RecurseSubPackages {
			return err
		}
		// print error message and continue if RecurseSubPackages is true
		fmt.Fprintf(w, "%s\n", err.Error())
	} else {
		fmt.Fprintf(w, "set %d field(s) of setter %q to value %q\n", count, r.Set.Name, r.Set.Value)
	}
	return nil
}
