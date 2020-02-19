// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package configcobra provides a target for embedding the config command group in another
// cobra command.
package configcobra

import (
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/api"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/tutorials"
	"sigs.k8s.io/kustomize/kyaml/commandutil"
)

var root = &cobra.Command{
	Use:   "config",
	Short: "[Alpha] Utilities for working with Resource Configuration.",
	Long: `[Alpha] Utilities for working with Resource Configuration.

Tutorials:

  Run 'kustomize help config tutorial-TUTORIAL'

	$ kustomize help config tutorials-command-basics

Command Documentation:

  Run 'kustomize help config CMD'

	$ kustomize help config tree

Advanced Documentation Topics:

  Run 'kustomize help config docs-TOPIC'

	$ kustomize help config docs-merge
	$ kustomize help config docs-merge3
	$ kustomize help config docs-fn
	$ kustomize help config docs-io-annotations
`,
}

// Export commands publicly for composition
var (
	Annotate           = commands.AnnotateCommand
	Cat                = commands.CatCommand
	Count              = commands.CountCommand
	CreateSetter       = commands.CreateSetterCommand
	CreateSubstitution = commands.CreateSubstitutionCommand
	Fmt                = commands.FmtCommand
	Grep               = commands.GrepCommand
	ListSetters        = commands.ListSettersCommand
	Merge              = commands.MergeCommand
	Merge3             = commands.Merge3Command
	RunFn              = commands.RunFnCommand
	Set                = commands.SetCommand
	Sink               = commands.SinkCommand
	Source             = commands.SourceCommand
	Tree               = commands.TreeCommand
	Wrap               = commands.WrapCommand
	XArgs              = commands.XArgsCommand

	StackOnError = &commands.StackOnError
	ExitOnError  = &commands.ExitOnError
)

// NewConfigCommand returns a new *cobra.Command for the config command group.  This may
// be embedded into other go binaries as a way of packaging the "config" command as part
// of another binary.
//
// name is substituted into the built-in documentation for each sub-command as the command
// invocation prefix -- e.g. if the result is embedded in kustomize, then name should be
// "kustomize" and the built-in docs will display "kustomize config" in the examples.
//
func NewConfigCommand(name string) *cobra.Command {
	// config command is alpha
	root.Version = "v0.0.0"

	// Only populate the command if Alpha commands are enabled.
	if !commandutil.GetAlphaEnabled() {
		// return the command because other subcommands are added to it
		root.Short = "[Alpha] To enable set KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true"
		root.Long = "[Alpha] To enable set KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true"
		root.Example = ""
		return root
	}

	root.PersistentFlags().BoolVar(&commands.StackOnError, "stack-trace", false,
		"print a stack-trace on failure")

	name = strings.TrimSpace(name + " config")
	commands.ExitOnError = true
	root.AddCommand(commands.AnnotateCommand(name))
	root.AddCommand(commands.GrepCommand(name))
	root.AddCommand(commands.TreeCommand(name))
	root.AddCommand(commands.CatCommand(name))
	root.AddCommand(commands.FmtCommand(name))
	root.AddCommand(commands.MergeCommand(name))
	root.AddCommand(commands.Merge3Command(name))
	root.AddCommand(commands.CountCommand(name))
	root.AddCommand(commands.RunFnCommand(name))
	root.AddCommand(commands.XArgsCommand())
	root.AddCommand(commands.WrapCommand())

	root.AddCommand(commands.SetCommand(name))
	root.AddCommand(commands.ListSettersCommand(name))
	root.AddCommand(commands.CreateSetterCommand(name))
	root.AddCommand(commands.CreateSubstitutionCommand(name))
	root.AddCommand(commands.SinkCommand(name))
	root.AddCommand(commands.SourceCommand(name))

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
