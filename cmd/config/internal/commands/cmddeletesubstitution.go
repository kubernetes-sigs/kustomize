// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewDeleteRunner returns a command runner.
func NewDeleteSubstitutionRunner(parent string) *DeleteSubstitutionRunner {
	r := &DeleteSubstitutionRunner{}
	c := &cobra.Command{
		Use:     "delete-subst DIR NAME",
		Args:    cobra.ExactArgs(2),
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	c.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"deletes substitution recursively in all the nested subpackages")
	runner.FixDocs(parent, c)
	r.Command = c

	return r
}

func DeleteSubstitutionCommand(parent string) *cobra.Command {
	return NewDeleteSubstitutionRunner(parent).Command
}

type DeleteSubstitutionRunner struct {
	Command            *cobra.Command
	DeleteSubstitution settersutil.DeleterCreator
	OpenAPIFile        string
	RecurseSubPackages bool
}

func (r *DeleteSubstitutionRunner) preRunE(c *cobra.Command, args []string) error {
	r.DeleteSubstitution.Name = args[1]
	r.DeleteSubstitution.DefinitionPrefix = fieldmeta.SubstitutionDefinitionPrefix

	r.OpenAPIFile = filepath.Join(args[0], ext.KRMFileName())

	return nil
}

func (r *DeleteSubstitutionRunner) runE(c *cobra.Command, args []string) error {
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

func (r *DeleteSubstitutionRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	sc, err := openapi.SchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName()))
	if err != nil {
		return err
	}
	r.DeleteSubstitution = settersutil.DeleterCreator{
		Name:               r.DeleteSubstitution.Name,
		DefinitionPrefix:   fieldmeta.SubstitutionDefinitionPrefix,
		RecurseSubPackages: r.RecurseSubPackages,
		OpenAPIFileName:    ext.KRMFileName(),
		OpenAPIPath:        filepath.Join(pkgPath, ext.KRMFileName()),
		ResourcesPath:      pkgPath,
		SettersSchema:      sc,
	}

	err = r.DeleteSubstitution.Delete()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.DeleteSubstitution.RecurseSubPackages {
			return err
		}
		// print error message and continue if RecurseSubPackages is true
		fmt.Fprintf(w, "%s\n", err.Error())
	} else {
		fmt.Fprintf(w, "deleted substitution %q\n", r.DeleteSubstitution.Name)
	}
	return nil
}
