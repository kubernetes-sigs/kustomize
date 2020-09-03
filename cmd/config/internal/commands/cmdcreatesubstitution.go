// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/pathutil"
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
	fixDocs(parent, cs)
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
	openAPIFileName, err := ext.OpenAPIFileName()
	if err != nil {
		return err
	}

	r.CreateSubstitution.OpenAPIFileName = openAPIFileName
	resourcePackagesPaths, err := pathutil.DirsWithFile(args[0], openAPIFileName, r.CreateSubstitution.RecurseSubPackages)
	if err != nil {
		return err
	}

	if len(resourcePackagesPaths) == 0 {
		return errors.Errorf("unable to find %q in package %q", r.CreateSubstitution.OpenAPIFileName, args[0])
	}

	for _, resourcesPath := range resourcePackagesPaths {
		r.CreateSubstitution = settersutil.SubstitutionCreator{
			Name:               r.CreateSubstitution.Name,
			FieldName:          r.CreateSubstitution.FieldName,
			FieldValue:         r.CreateSubstitution.FieldValue,
			RecurseSubPackages: r.CreateSubstitution.RecurseSubPackages,
			Pattern:            r.CreateSubstitution.Pattern,
			OpenAPIFileName:    openAPIFileName,
			OpenAPIPath:        filepath.Join(resourcesPath, openAPIFileName),
			ResourcesPath:      resourcesPath,
		}

		// Add schema present in openAPI file for current package
		if err := openapi.AddSchemaFromFile(r.CreateSubstitution.OpenAPIPath); err != nil {
			return err
		}
		err := r.CreateSubstitution.Create()
		if err != nil {
			// return err if there is only package
			if len(resourcePackagesPaths) == 1 {
				return err
			} else {
				// print error message and continue if there are multiple packages to set
				fmt.Fprintf(c.OutOrStdout(), "%s in package %q\n\n", err.Error(), r.CreateSubstitution.ResourcesPath)
			}
		} else {
			fmt.Fprintf(c.OutOrStdout(), "created substitution %q in package %q\n\n", r.CreateSubstitution.Name, r.CreateSubstitution.ResourcesPath)
		}

		// Delete schema present in openAPI file for current package
		if err := openapi.DeleteSchemaInFile(r.CreateSubstitution.OpenAPIPath); err != nil {
			return err
		}
	}
	return nil
}

func (r *CreateSubstitutionRunner) preRun(c *cobra.Command, args []string) {
	r.CreateSubstitution.Name = args[1]
}
