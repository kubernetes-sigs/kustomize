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
)

var root = &cobra.Command{
	Use:     "config",
	Short:   "Utilities for working with Resource Configuration.",
	Long:    `Utilities for working with Resource Configuration.`,
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
		Short: "Documentation for merging Resources (2-way merge).",
		Long:  api.Merge2Long,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-merge3",
		Short: "Documentation for merging Resources (3-way merge).",
		Long:  api.Merge3Long,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-fn",
		Short: "Documentation for writing containerized functions run by run-fns.",
		Long:  api.ConfigFnLong,
	})
	root.AddCommand(&cobra.Command{
		Use:   "docs-io-annotations",
		Short: "Documentation for annotations used by io.",
		Long:  api.ConfigIoLong,
	})
	return root
}
