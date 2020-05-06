// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewCreateSubstitutionRunner returns a command runner.
func NewCreateSubstitutionRunner(parent string) *CreateSubstitutionRunner {
	r := &CreateSubstitutionRunner{}
	cs := &cobra.Command{
		Use:     "create-subst DIR NAME",
		Args:    cobra.ExactArgs(2),
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	cs.Flags().StringVar(&r.CreateSubstitution.FieldName, "field", "",
		"name of the field to set -- e.g. --field image")
	cs.Flags().StringVar(&r.CreateSubstitution.FieldValue, "field-value", "",
		"value of the field to create substitution for -- e.g. --field-value nginx:0.1.0")
	cs.Flags().StringVar(&r.CreateSubstitution.Pattern, "pattern", "",
		`substitution pattern -- e.g. --pattern \${my-image-setter}:\${my-tag-setter}`)
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
	return handleError(c, r.CreateSubstitution.Create(r.OpenAPIFile, args[0]))
}

func (r *CreateSubstitutionRunner) preRunE(c *cobra.Command, args []string) error {
	var err error
	r.CreateSubstitution.Name = args[1]
	if err != nil {
		return err
	}

	r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}

	if err := openapi.AddSchemaFromFile(r.OpenAPIFile); err != nil {
		return err
	}

	// check if setter with same name exists and throw error
	ref, err := spec.NewRef(setters2.DefinitionsPrefix + setters2.SetterDefinitionPrefix + r.CreateSubstitution.Name)
	if err != nil {
		return err
	}

	setter, _ := openapi.Resolve(&ref)
	// if setter already exists with input substitution name, throw error
	if setter != nil {
		return errors.Errorf(fmt.Sprintf("setter with name %s already exists, "+
			"substitution and setter can't have same name", r.CreateSubstitution.Name))
	}

	// extract setter name tokens from pattern enclosed in ${}
	re := regexp.MustCompile(`\$\{([^}]*)\}`)
	markers := re.FindAll([]byte(r.CreateSubstitution.Pattern), -1)
	if len(markers) == 0 {
		return errors.Errorf("unable to find setter names in pattern, " +
			"setter names must be enclosed in ${}")
	}

	for _, marker := range markers {
		ref := setters2.DefinitionsPrefix + setters2.SetterDefinitionPrefix +
			strings.TrimSuffix(strings.TrimPrefix(string(marker), "${"), "}")
		r.CreateSubstitution.Values = append(
			r.CreateSubstitution.Values,
			setters2.Value{Marker: string(marker), Ref: ref},
		)
	}

	return nil
}
