// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package cmds provides a target for embedding the config command group in another
// cobra command.
package cmds

import (
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/cmd"
	"sigs.k8s.io/kustomize/cmd/config/cmddocs/api"
	"sigs.k8s.io/kustomize/cmd/config/cmddocs/tutorials"
)

var root = &cobra.Command{
	Use:   "config",
	Short: "[Alpha] Utilities for working with Resource Configuration.",
	Long: `[Alpha] Utilities for working with Resource Configuration.

Tutorials:

  Run 'kustomize help config tutorial-TUTORIAL'

	$ kustomize help config tutorials-basics

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
	Version: "v0.0.1",
}

// NewConfigCommand returns a new *cobra.Command for the config command group.  This may
// be embedded into other go binaries as a way of packaging the "config" command as part
// of another binary.
//
// name is substituted into the built-in documentation for each sub-command as the command
// invocation prefix -- e.g. if the result is embedded in kustomize, then name should be
// "kustomize" and the built-in docs will display "kustomize config" in the examples.
//
func NewConfigCommand(name string) *cobra.Command {
	root.PersistentFlags().BoolVar(&cmd.StackOnError, "stack-trace", false,
		"print a stack-trace on failure")

	name = strings.TrimSpace(name + " config")
	cmd.ExitOnError = true
	root.AddCommand(cmd.GrepCommand(name))
	root.AddCommand(cmd.TreeCommand(name))
	root.AddCommand(cmd.CatCommand(name))
	root.AddCommand(cmd.FmtCommand(name))
	root.AddCommand(cmd.MergeCommand(name))
	root.AddCommand(cmd.CountCommand(name))
	root.AddCommand(cmd.RunFnCommand(name))

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
		Short: "[Alpha] Documentation for writing containerized functions run by run.",
		Long:  api.ConfigFnLong,
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
