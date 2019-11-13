// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/cmd"
	"sigs.k8s.io/kustomize/cmd/config/cmddocs/api"
)

//go:generate $GOBIN/mdtogo docs/api-conventions cmddocs/api --full=true
//go:generate $GOBIN/mdtogo docs/commands cmddocs/commands
var root = &cobra.Command{
	Use:   "config",
	Short: "Utilities for working with Resource Configuration.",
	Long:  `Utilities for working with Resource Configuration.`,
}

func main() {
	root.PersistentFlags().BoolVar(&cmd.StackOnError, "stack-trace", false,
		"print a stack-trace on failure")

	name := "config"
	cmd.ExitOnError = true
	root.AddCommand(cmd.GrepCommand(name))
	root.AddCommand(cmd.TreeCommand(name))
	root.AddCommand(cmd.CatCommand(name))
	root.AddCommand(cmd.FmtCommand(name))
	root.AddCommand(cmd.MergeCommand(name))
	root.AddCommand(cmd.CountCommand(name))
	root.AddCommand(cmd.RunFnCommand(name))

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

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
