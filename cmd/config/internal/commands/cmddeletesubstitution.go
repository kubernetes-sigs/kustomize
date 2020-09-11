// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
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
	fixDocs(parent, c)
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
	var err error
	r.DeleteSubstitution.Name = args[1]
	r.DeleteSubstitution.DefinitionPrefix = fieldmeta.SubstitutionDefinitionPrefix

	r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}

	return nil
}

func (r *DeleteSubstitutionRunner) runE(c *cobra.Command, args []string) error {
	e := executeCmdOnPkgs{
		needOpenAPI:        true,
		writer:             c.OutOrStdout(),
		rootPkgPath:        args[0],
		recurseSubPackages: r.RecurseSubPackages,
		cmdRunner:          r,
	}
	err := e.execute()
	if err != nil {
		return handleError(c, err)
	}
	return nil
}

func (r *DeleteSubstitutionRunner) executeCmd(w io.Writer, pkgPath string) error {
	openAPIFileName, err := ext.OpenAPIFileName()
	if err != nil {
		return err
	}
	r.DeleteSubstitution = settersutil.DeleterCreator{
		Name:               r.DeleteSubstitution.Name,
		DefinitionPrefix:   fieldmeta.SubstitutionDefinitionPrefix,
		RecurseSubPackages: r.RecurseSubPackages,
		OpenAPIFileName:    openAPIFileName,
		OpenAPIPath:        filepath.Join(pkgPath, openAPIFileName),
		ResourcesPath:      pkgPath,
	}

	err = r.DeleteSubstitution.Delete()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.DeleteSubstitution.RecurseSubPackages {
			return err
		} else {
			// print error message and continue if RecurseSubPackages is true
			fmt.Fprintf(w, "%s\n", err.Error())
		}
	} else {
		fmt.Fprintf(w, "deleted substitution %q\n", r.DeleteSubstitution.Name)
	}
	return nil
}
