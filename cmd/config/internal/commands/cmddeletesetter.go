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
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewDeleteRunner returns a command runner.
func NewDeleteSetterRunner(parent string) *DeleteSetterRunner {
	r := &DeleteSetterRunner{}
	c := &cobra.Command{
		Use:     "delete-setter DIR NAME",
		Args:    cobra.ExactArgs(2),
		Short:   commands.DeleteSetterShort,
		Long:    commands.DeleteSetterLong,
		Example: commands.DeleteSetterExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	c.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"deletes setter recursively in all the nested subpackages")
	runner.FixDocs(parent, c)
	r.Command = c

	return r
}

func DeleteSetterCommand(parent string) *cobra.Command {
	return NewDeleteSetterRunner(parent).Command
}

type DeleteSetterRunner struct {
	Command            *cobra.Command
	DeleteSetter       settersutil.DeleterCreator
	OpenAPIFile        string
	RecurseSubPackages bool
}

func (r *DeleteSetterRunner) preRunE(c *cobra.Command, args []string) error {
	r.DeleteSetter.Name = args[1]
	r.DeleteSetter.DefinitionPrefix = fieldmeta.SetterDefinitionPrefix

	r.OpenAPIFile = filepath.Join(args[0], ext.KRMFileName())

	return nil
}

func (r *DeleteSetterRunner) runE(c *cobra.Command, args []string) error {
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

func (r *DeleteSetterRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	sc, err := openapi.SchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName()))
	if err != nil {
		return err
	}
	r.DeleteSetter = settersutil.DeleterCreator{
		Name:               r.DeleteSetter.Name,
		DefinitionPrefix:   fieldmeta.SetterDefinitionPrefix,
		RecurseSubPackages: r.RecurseSubPackages,
		OpenAPIFileName:    ext.KRMFileName(),
		OpenAPIPath:        filepath.Join(pkgPath, ext.KRMFileName()),
		ResourcesPath:      pkgPath,
		SettersSchema:      sc,
	}

	err = r.DeleteSetter.Delete()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.DeleteSetter.RecurseSubPackages {
			return err
		}
		// print error message and continue if RecurseSubPackages is true
		fmt.Fprintf(w, "%s\n", err.Error())
	} else {
		fmt.Fprintf(w, "deleted setter %q\n", r.DeleteSetter.Name)
	}
	return nil
}
