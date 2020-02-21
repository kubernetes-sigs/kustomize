// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewCreateSubstitutionRunner returns a command runner.
func NewCreateSubstitutionRunner(parent string) *CreateSubstitutionRunner {
	r := &CreateSubstitutionRunner{}
	cs := &cobra.Command{
		Use:     "create-subst DIR NAME VALUE",
		Args:    cobra.ExactArgs(3),
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	cs.Flags().StringVar(&r.CreateSubstitution.FieldName, "field", "",
		"name of the field to set -- e.g. --field port")
	cs.Flags().StringVar(&r.CreateSubstitution.Pattern, "pattern", "",
		"substitution pattern")
	cs.Flags().StringSliceVar(&r.Values, "value", []string{""},
		"substitution values for the pattern.  format is PATTERN_MARKER=SETTER_NAME"+
			"where PATTERN_MARKER is the pattern substring to replace, and SETTER_NAME is the"+
			"setter from which to take the replacement value.")
	_ = cs.MarkFlagRequired("pattern")
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
	r.CreateSubstitution.FieldValue = args[2]
	if err != nil {
		return err
	}

	r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}

	// parse the marker values
	for i := range r.Values {
		parts := strings.SplitN(r.Values[i], "=", 2)
		if len(parts) < 2 {
			return errors.Errorf("values must be specified as PATTERN_MARKER=SETTER_NAME")
		}
		ref := setters2.DefinitionsPrefix + setters2.SetterDefinitionPrefix + parts[1]
		r.CreateSubstitution.Values = append(
			r.CreateSubstitution.Values,
			setters2.Value{Marker: parts[0], Ref: ref},
		)
	}

	return nil
}
