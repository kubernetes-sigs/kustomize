// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/go-openapi/spec"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/setters"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewCreateSetterRunner returns a command runner.
func NewCreateSetterRunner(parent string) *CreateSetterRunner {
	r := &CreateSetterRunner{}
	set := &cobra.Command{
		Use:     "create-setter DIR NAME VALUE",
		Args:    cobra.RangeArgs(2, 3),
		Short:   commands.CreateSetterShort,
		Long:    commands.CreateSetterLong,
		Example: commands.CreateSetterExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	set.Flags().StringVar(&r.Set.SetPartialField.Setter.Value, "value", "",
		"optional flag, alternative to specifying the value as an argument. e.g. used to specify values that start with '-'")
	set.Flags().StringVar(&r.Set.SetPartialField.SetBy, "set-by", "",
		"record who the field was default by.")
	set.Flags().StringVar(&r.Set.SetPartialField.Description, "description", "",
		"record a description for the current setter value.")
	set.Flags().StringVar(&r.Set.SetPartialField.Field, "field", "",
		"name of the field to set -- e.g. --field port.  defaults to all fields match"+
			"VALUE.  maybe be the field name, field path, or partial field path (suffix)")
	set.Flags().StringVar(&r.Set.ResourceMeta.Name, "name", "",
		"name of the Resource on which to create the setter.")
	set.Flags().MarkHidden("name")
	set.Flags().StringVar(&r.Set.ResourceMeta.Kind, "kind", "",
		"kind of the Resource on which to create the setter.")
	set.Flags().MarkHidden("kind")
	set.Flags().StringVar(&r.Set.SetPartialField.Type, "type", "",
		"OpenAPI field type for the setter -- e.g. integer,boolean,string.")
	set.Flags().BoolVar(&r.Set.SetPartialField.Partial, "partial", false,
		"create a partial setter for only part of the field value.")
	set.Flags().MarkHidden("partial")
	set.Flags().StringVar(&setterVersion, "version", "",
		"use this version of the setter format")
	set.Flags().BoolVar(&r.CreateSetter.Required, "required", false,
		"indicates that this setter must be set by package consumer before live apply/preview")
	set.Flags().StringVar(&r.CreateSetter.SchemaPath, "schema-path", "",
		`openAPI schema file path for setter constraints -- file content `+
			`e.g. {"type": "string", "maxLength": 15, "enum": ["allowedValue1", "allowedValue2"]}`)
	set.Flags().MarkHidden("version")
	fixDocs(parent, set)
	r.Command = set
	return r
}

func CreateSetterCommand(parent string) *cobra.Command {
	return NewCreateSetterRunner(parent).Command
}

type CreateSetterRunner struct {
	Command      *cobra.Command
	Set          setters.CreateSetter
	CreateSetter settersutil.SetterCreator
	OpenAPIFile  string
}

func (r *CreateSetterRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, r.set(c, args))
}

func (r *CreateSetterRunner) preRunE(c *cobra.Command, args []string) error {
	valueSetFromFlag := c.Flag("value").Changed
	var err error
	r.Set.SetPartialField.Setter.Name = args[1]
	r.CreateSetter.Name = args[1]
	if valueSetFromFlag {
		r.CreateSetter.FieldValue = r.Set.SetPartialField.Setter.Value
	} else if len(args) > 2 {
		r.Set.SetPartialField.Setter.Value = args[2]
		r.CreateSetter.FieldValue = args[2]
	}
	r.CreateSetter.FieldName, err = c.Flags().GetString("field")
	if err != nil {
		return err
	}

	if setterVersion == "" {
		if len(args) == 2 && r.Set.SetPartialField.Type == "array" && c.Flag("field").Changed {
			setterVersion = "v2"
		} else if len(args) < 2 || !c.Flag("value").Changed && len(args) < 3 {
			setterVersion = "v1"
		} else if err := initSetterVersion(c, args); err != nil {
			return err
		}
	}
	if setterVersion == "v2" {
		var err error
		r.OpenAPIFile, err = ext.GetOpenAPIFile(args)
		if err != nil {
			return err
		}

		if err := openapi.AddSchemaFromFile(r.OpenAPIFile); err != nil {
			return err
		}

		// check if substitution with same name exists and throw error
		ref, err := spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + r.CreateSetter.Name)
		if err != nil {
			return err
		}

		subst, _ := openapi.Resolve(&ref)
		// if substitution already exists with the input setter name, throw error
		if subst != nil {
			return errors.Errorf("substitution with name %s already exists, "+
				"substitution and setter can't have same name", r.CreateSetter.Name)
		}

		r.CreateSetter.Description = r.Set.SetPartialField.Description
		r.CreateSetter.SetBy = r.Set.SetPartialField.SetBy
		r.CreateSetter.Type = r.Set.SetPartialField.Type

		if r.CreateSetter.Type == "array" {
			if !c.Flag("field").Changed {
				return errors.Errorf("field flag must be set for array type setters")
			}
		}
	}
	return nil
}

func (r *CreateSetterRunner) set(c *cobra.Command, args []string) error {
	if setterVersion == "v2" {
		return r.CreateSetter.Create(r.OpenAPIFile, args[0])
	}

	rw := &kio.LocalPackageReadWriter{PackagePath: args[0]}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&r.Set},
		Outputs: []kio.Writer{rw}}.Execute()
	if err != nil {
		return err
	}
	return nil
}
