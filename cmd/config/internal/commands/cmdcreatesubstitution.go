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
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewCreateSubstitutionRunner returns a command runner.
func NewCreateSubstitutionRunner(parent string) *CreateSubstitutionRunner {
	r := &CreateSubstitutionRunner{}
	cs := &cobra.Command{
		Use:    "create-subst DIR NAME",
		Args:   cobra.ExactArgs(2),
		PreRun: r.preRun,
		RunE:   r.runE,
		Deprecated: "imperative substitutions will no longer be available in kustomize v5.\n" +
			"See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.",
	}
	cs.Flags().StringVar(&r.CreateSubstitution.FieldName, "field", "",
		"name of the field to set -- e.g. --field image")
	cs.Flags().StringVar(&r.CreateSubstitution.FieldValue, "field-value", "",
		"value of the field to create substitution for -- e.g. --field-value nginx:0.1.0")
	cs.Flags().StringVar(&r.CreateSubstitution.Pattern, "pattern", "",
		`substitution pattern -- e.g. --pattern \${my-image-setter}:\${my-tag-setter}`)
	cs.Flags().BoolVarP(&r.CreateSubstitution.RecurseSubPackages, "recurse-subpackages", "R", false,
		"creates substitution recursively in all the nested subpackages")
	_ = cs.MarkFlagRequired("pattern")
	_ = cs.MarkFlagRequired("field-value")
	runner.FixDocs(parent, cs)
	r.Command = cs
	return r
}

func CreateSubstitutionCommand(parent string) *cobra.Command {
	return NewCreateSubstitutionRunner(parent).Command
}

type CreateSubstitutionRunner struct {
	Command            *cobra.Command
	CreateSubstitution settersutil.SubstitutionCreator
	OpenAPIFile        string
	Values             []string
}

func (r *CreateSubstitutionRunner) runE(c *cobra.Command, args []string) error {
	e := runner.ExecuteCmdOnPkgs{
		NeedOpenAPI:        true,
		Writer:             c.OutOrStdout(),
		RootPkgPath:        args[0],
		RecurseSubPackages: r.CreateSubstitution.RecurseSubPackages,
		CmdRunner:          r,
	}
	err := e.Execute()
	if err != nil {
		return runner.HandleError(c, err)
	}

	return nil
}

func (r *CreateSubstitutionRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	sc, err := openapi.SchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName()))
	if err != nil {
		return err
	}
	r.CreateSubstitution = settersutil.SubstitutionCreator{
		Name:               r.CreateSubstitution.Name,
		FieldName:          r.CreateSubstitution.FieldName,
		FieldValue:         r.CreateSubstitution.FieldValue,
		RecurseSubPackages: r.CreateSubstitution.RecurseSubPackages,
		Pattern:            r.CreateSubstitution.Pattern,
		OpenAPIFileName:    ext.KRMFileName(),
		OpenAPIPath:        filepath.Join(pkgPath, ext.KRMFileName()),
		ResourcesPath:      pkgPath,
		SettersSchema:      sc,
	}

	err = r.CreateSubstitution.Create()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.CreateSubstitution.RecurseSubPackages {
			return err
		}
		// print error message and continue if RecurseSubPackages is true
		fmt.Fprintf(w, "%s\n", err.Error())
	} else {
		fmt.Fprintf(w, "created substitution %q\n", r.CreateSubstitution.Name)
	}
	return nil
}

func (r *CreateSubstitutionRunner) preRun(c *cobra.Command, args []string) {
	r.CreateSubstitution.Name = args[1]
}
