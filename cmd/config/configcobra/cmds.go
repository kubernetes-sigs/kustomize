// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package configcobra provides a target for embedding the config command group in another
// cobra command.
package configcobra

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/api"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/tutorials"
	"sigs.k8s.io/kustomize/cmd/config/runner"
)

// Export commands publicly for composition
//
//nolint:gochecknoglobals
var (
	Cat   = commands.CatCommand
	Count = commands.CountCommand
	Grep  = commands.GrepCommand
	RunFn = commands.RunCommand
	Tree  = commands.TreeCommand

	StackOnError = &runner.StackOnError
	ExitOnError  = &runner.ExitOnError
)

// AddCommands adds the cfg and fn commands to kustomize.
func AddCommands(root *cobra.Command, name string) *cobra.Command {
	runner.ExitOnError = true

	root.PersistentFlags().BoolVar(StackOnError, "stack-trace", false,
		"print a stack-trace on error")

	root.AddCommand(GetCfg(name))
	root.AddCommand(GetFn(name))

	root.AddCommand(&cobra.Command{
		Use:   "docs-merge",
		Short: "[Alpha] Documentation for merging Resources (2-way merge).",
		Long:  api.Merge2Long,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-merge3",
		Short: "[Alpha] Documentation for merging Resources (3-way merge).",
		Long:  api.Merge3Long,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-fn",
		Short: "[Alpha] Documentation for developing and invoking Configuration Functions.",
		Long:  api.FunctionsImplLong,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-fn-spec",
		Short: "[Alpha] Documentation for Configuration Functions Specification.",
		Long:  api.FunctionsSpecLong,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-io-annotations",
		Short: "[Alpha] Documentation for annotations used by io.",
		Long:  api.ConfigIoLong,
	})

	root.AddCommand(&cobra.Command{
		Use:   "tutorials-command-basics",
		Short: "[Alpha] Tutorials for using basic config commands.",
		Long:  tutorials.ConfigurationBasicsLong,
	})

	root.AddCommand(&cobra.Command{
		Use:   "tutorials-function-basics",
		Short: "[Alpha] Tutorials for using functions.",
		Long:  tutorials.FunctionBasicsLong,
	})

	return root
}
