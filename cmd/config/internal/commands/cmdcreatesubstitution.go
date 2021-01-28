// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-openapi/spec"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
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
	cs.Flags().StringVar(&r.FieldName, "field", "",
		"name of the field to set -- e.g. --field image")
	cs.Flags().StringVar(&r.FieldValue, "field-value", "",
		"value of the field to create substitution for -- e.g. --field-value nginx:0.1.0")
	cs.Flags().StringVar(&r.Pattern, "pattern", "",
		`substitution pattern -- e.g. --pattern \${my-image-setter}:\${my-tag-setter}`)
	cs.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"creates substitution recursively in all the nested subpackages")
	cs.Flags().BoolVar(&r.AllSubst, "all-subst", false,
		"creates all possible substitutions for field values which contain setter values")
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
	Name               string
	FieldName          string
	FieldValue         string
	Pattern            string
	RecurseSubPackages bool
	OpenAPIFile        string
	Values             []string
	AllSubst           bool
}

func (r *CreateSubstitutionRunner) runE(c *cobra.Command, args []string) error {
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

func (r *CreateSubstitutionRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	sc, err := openapi.SchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName()))
	if err != nil {
		return err
	}

	substInfos, err := r.substInfos(pkgPath, sc)
	if err != nil {
		return err
	}
	if len(substInfos) == 0 {
		return nil
	}
	for i, substInfo := range substInfos {
		name := r.Name
		if r.AllSubst {
			name = fmt.Sprintf("%s_%d", name, i+1)
		}
		r.CreateSubstitution = settersutil.SubstitutionCreator{
			Name:               name,
			FieldName:          r.FieldName,
			FieldValue:         substInfo.FieldValue,
			RecurseSubPackages: r.RecurseSubPackages,
			Pattern:            substInfo.Pattern,
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
			} else {
				// print error message and continue if RecurseSubPackages is true
				fmt.Fprintf(w, "%s\n", err.Error())
			}
		} else {
			fmt.Fprintf(w, "created substitution %q\n", r.CreateSubstitution.Name)
		}
	}
	return nil
}

func (r *CreateSubstitutionRunner) preRunE(c *cobra.Command, args []string) error {
	r.Name = args[1]
	if !r.AllSubst {
		if !c.Flag("field-value").Changed {
			return errors.Errorf(`"field-value" must be specified`)
		}

		if !c.Flag("pattern").Changed {
			return errors.Errorf(`"pattern" must be specified`)
		}
	}
	return nil
}

// substInfos returns the information regarding substitutions to be created
// it simply returns the user provided info if "--all-subst" flag is false
// else returns information regarding all the possible substitutions which can be
// created using available setters in the package
func (r *CreateSubstitutionRunner) substInfos(pkgPath string, sc *spec.Schema) ([]setters2.SubstInfo, error) {
	var substInfos []setters2.SubstInfo
	if r.AllSubst {
		s := setters2.SubstUtil{
			SetterInfos: setterInfos(sc),
		}
		in := &kio.LocalPackageReader{PackagePath: pkgPath, PackageFileName: ext.KRMFileName()}
		err := kio.Pipeline{
			Inputs:  []kio.Reader{in},
			Filters: []kio.Filter{kio.FilterAll(&s)},
		}.Execute()
		if err != nil {
			return nil, err
		}
		substInfos = s.SubstInfos
	} else {
		// simply return the user provided information
		substInfos = append(substInfos, setters2.SubstInfo{
			FieldValue: r.FieldValue,
			Pattern:    r.Pattern,
		})
	}
	return substInfos, nil
}

// setterInfos returns the scalar setters information present in
// the provided setter schema from the package
func setterInfos(sc *spec.Schema) []setters2.SetterInfo {
	if sc == nil {
		return nil
	}
	var setterInfos []setters2.SetterInfo
	for _, schema := range sc.Definitions {
		ext, err := setters2.GetExtFromSchema(&schema)
		if err != nil {
			return nil
		}
		if ext.Setter != nil && ext.Setter.Value != "" {
			setterInfos = append(setterInfos, setters2.SetterInfo{
				SetterName:  ext.Setter.Name,
				SetterValue: ext.Setter.Value,
			})
		}
	}
	return setterInfos
}
