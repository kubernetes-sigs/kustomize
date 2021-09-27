// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
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
		Deprecated: "setter commands will no longer be available in kustomize v5.\n" +
			"See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.",
	}
	set.Flags().StringVar(&r.FieldValue, "value", "",
		"optional flag, alternative to specifying the value as an argument. e.g. used to specify values that start with '-'")
	set.Flags().StringVar(&r.SetBy, "set-by", "",
		"record who the field was default by.")
	set.Flags().StringVar(&r.Description, "description", "",
		"record a description for the current setter value.")
	set.Flags().StringVar(&r.FieldName, "field", "",
		"name of the field to set, a suffix of the path to the field, or the full"+
			" path to the field. Default is to match all fields.")
	set.Flags().StringVar(&r.Type, "type", "",
		"OpenAPI field type for the setter -- e.g. integer,boolean,string.")
	set.Flags().BoolVar(&r.Required, "required", false,
		"indicates that this setter must be set by package consumer before live apply/preview")
	set.Flags().StringVar(&r.SchemaPath, "schema-path", "",
		`openAPI schema file path for setter constraints -- file content `+
			`e.g. {"type": "string", "maxLength": 15, "enum": ["allowedValue1", "allowedValue2"]}`)
	set.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"creates setter recursively in all the nested subpackages")
	set.Flags().MarkHidden("version")
	runner.FixDocs(parent, set)
	r.Command = set
	return r
}

func CreateSetterCommand(parent string) *cobra.Command {
	return NewCreateSetterRunner(parent).Command
}

type CreateSetterRunner struct {
	Command            *cobra.Command
	CreateSetter       settersutil.SetterCreator
	OpenAPIFile        string
	SchemaPath         string
	FieldValue         string
	SetBy              string
	Description        string
	SetterName         string
	Type               string
	FieldName          string
	Schema             string
	Required           bool
	RecurseSubPackages bool
}

func (r *CreateSetterRunner) runE(c *cobra.Command, args []string) error {
	return runner.HandleError(c, r.createSetter(c, args))
}

func (r *CreateSetterRunner) preRunE(c *cobra.Command, args []string) error {
	var err error
	r.SetterName = args[1]
	if len(args) > 2 {
		r.FieldValue = args[2]
	}
	r.FieldName, err = c.Flags().GetString("field")
	if err != nil {
		return err
	}

	if r.Type != "array" && !c.Flag("value").Changed && len(args) < 3 {
		return errors.Errorf("setter name and value must be provided, " +
			"value can either be an argument or can be passed as a flag --value")
	}

	err = r.processSchema()
	if err != nil {
		return err
	}

	if r.Type == "array" {
		if !c.Flag("field").Changed {
			return errors.Errorf("field flag must be set for array type setters")
		}
	}
	return nil
}

func (r *CreateSetterRunner) processSchema() error {
	sc, err := schemaFromFile(r.SchemaPath)
	if err != nil {
		return err
	}

	flagType := r.Type
	var schemaType string
	switch {
	// json schema allows more than one type to be specified, but openapi
	// only allows one. So we follow the openapi convention.
	case len(sc.Type) > 1:
		return errors.Errorf("only one type is supported: %s",
			strings.Join(sc.Type, ", "))
	case len(sc.Type) == 1:
		schemaType = sc.Type[0]
	}

	// Since type can be set both through the schema file and through the
	// --type flag, we make sure the same value is set in both places. If they
	// are both set with different values, we return an error.
	switch {
	case flagType == "" && schemaType != "":
		r.Type = schemaType
	case flagType != "" && schemaType == "":
		sc.Type = []string{flagType}
	case flagType != "" && schemaType != "":
		if flagType != schemaType {
			return errors.Errorf("type provided in type flag (%s) and in schema (%s) doesn't match",
				r.Type, sc.Type[0])
		}
	}

	// Only marshal the properties in SchemaProps. This means any fields in
	// the schema file that isn't recognized will be dropped.
	// TODO: Consider if we should return an error here instead of just dropping
	// the unknown fields.
	b, err := json.Marshal(sc.SchemaProps)
	if err != nil {
		return errors.Errorf("error marshalling schema: %v", err)
	}
	r.Schema = string(b)
	return nil
}

func (r *CreateSetterRunner) createSetter(c *cobra.Command, args []string) error {
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

func (r *CreateSetterRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	sc, err := openapi.SchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName()))
	if err != nil {
		return err
	}
	r.CreateSetter = settersutil.SetterCreator{
		Name:               r.SetterName,
		SetBy:              r.SetBy,
		Description:        r.Description,
		Type:               r.Type,
		Schema:             r.Schema,
		FieldName:          r.FieldName,
		FieldValue:         r.FieldValue,
		Required:           r.Required,
		RecurseSubPackages: r.RecurseSubPackages,
		OpenAPIFileName:    ext.KRMFileName(),
		OpenAPIPath:        filepath.Join(pkgPath, ext.KRMFileName()),
		ResourcesPath:      pkgPath,
		SettersSchema:      sc,
	}

	err = r.CreateSetter.Create()
	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.CreateSetter.RecurseSubPackages {
			return err
		}
		// print error message and continue if RecurseSubPackages is true
		fmt.Fprintf(w, "%s\n", err.Error())
	} else {
		fmt.Fprintf(w, "created setter %q\n", r.CreateSetter.Name)
	}
	return nil
}

// schemaFromFile reads the contents from schemaPath and returns schema
func schemaFromFile(schemaPath string) (*spec.Schema, error) {
	sc := &spec.Schema{}
	if schemaPath == "" {
		return sc, nil
	}
	sch, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return sc, err
	}

	if len(sch) == 0 {
		return sc, nil
	}

	err = sc.UnmarshalJSON(sch)
	if err != nil {
		return sc, errors.Errorf("unable to parse schema: %v", err)
	}
	return sc, nil
}
