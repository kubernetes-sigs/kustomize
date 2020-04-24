// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/setters"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// NewUpgradeRunner returns a command runner.
func NewUpgradeRunner(parent string) *UpgradeRunner {
	r := &UpgradeRunner{}
	c := &cobra.Command{
		Use:     "upgrade DIR",
		Args:    cobra.MaximumNArgs(1),
		Short:   commands.UpgradeShort,
		Long:    commands.UpgradeLong,
		Example: commands.UpgradeExamples,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c
	return r
}

func UpgradeCommand(parent string) *cobra.Command {
	return NewUpgradeRunner(parent).Command
}

type UpgradeRunner struct {
	Command *cobra.Command
	Lookup  setters.LookupSettersDeprecated
}

func (r *UpgradeRunner) runE(c *cobra.Command, args []string) error {
	// list setters and upgrade them from resource files
	r.Lookup.Upgrade = true
	rw := &kio.LocalPackageReadWriter{
		PackagePath: args[0],
	}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&r.Lookup},
		Outputs: []kio.Writer{rw},
	}.Execute()
	if err != nil {
		return err
	}

	openAPIPath, err := ext.GetOpenAPIFile(args)
	if err != nil {
		return err
	}

	_, err = os.Stat(openAPIPath)
	if err != nil {
		return err
	}

	// for each v1 setter create the equivalent in v2, partial setters
	// can't be converted to substitutions due to not enough information,
	// they should be manually created by user
	upgradedCount := 0
	for _, setter := range r.Lookup.SetterCounts {
		sc := settersutil.SetterCreator{
			Name:          setter.Name,
			FieldValue:    setter.Value,
			Description:   setter.Description,
			SetBy:         setter.SetBy,
			Type:          setter.Type,
			ResourcesPath: args[0],
			OpenAPIPath:   openAPIPath,
		}
		err = kio.Pipeline{
			Inputs:  []kio.Reader{rw},
			Filters: []kio.Filter{&sc},
			Outputs: []kio.Writer{rw},
		}.Execute()
		if err != nil {
			fmt.Fprintf(c.OutOrStdout(), "encountered error: %s while migrating setter %s\n", err.Error(), setter.Name)
		} else {
			upgradedCount++
		}
	}
	fmt.Fprint(c.OutOrStdout(), "\n"+`if your resources contain partial setters, they are not automatically converted to substitutions,
however the references are removed from resources and constituent setters are created,
please use create-subst command to create equivalent substitutions using upgraded setters`+"\n\n")

	fmt.Fprintf(c.OutOrStdout(), "upgraded %d setters\n", upgradedCount)
	return nil
}
