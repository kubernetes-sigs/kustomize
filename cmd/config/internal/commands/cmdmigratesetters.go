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

// NewMigrateSettersRunner returns a command runner.
func NewMigrateSettersRunner(parent string) *MigrateSettersRunner {
	r := &MigrateSettersRunner{}
	c := &cobra.Command{
		Use:     "migrate-setters DIR",
		Args:    cobra.MaximumNArgs(1),
		Short:   commands.MigrateSettersShort,
		Long:    commands.MigrateSettersLong,
		Example: commands.MigrateSettersExamples,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c
	return r
}

func MigrateSettersCommand(parent string) *cobra.Command {
	return NewMigrateSettersRunner(parent).Command
}

type MigrateSettersRunner struct {
	Command *cobra.Command
	Lookup  setters.LookupDeleteSetters
}

func (r *MigrateSettersRunner) runE(c *cobra.Command, args []string) error {
	// list v1 setters and delete them from resource files
	r.Lookup.Delete = true
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

	// for each v1 setter, create the equivalent in v2 partial setters
	// can't be converted to substitutions due to not enough information,
	// they should be manually created by user
	migratedCount := 0
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
			migratedCount++
		}
	}
	fmt.Fprintf(c.OutOrStdout(), "migrated %d setters\n", migratedCount)
	return nil
}
