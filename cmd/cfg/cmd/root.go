// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge3"
)

func GetRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "cfg",
		Short: "Manage Resource configuration files",
		Long:  `Manage Resource configuration files`,
	}
	root.PersistentFlags().BoolVar(&StackOnError, "stack-trace", false,
		"print a stack-trace on failure")

	ExitOnError = true
	root.AddCommand(GrepCommand())
	root.AddCommand(TreeCommand())
	root.AddCommand(CatCommand())
	root.AddCommand(FmtCommand())
	root.AddCommand(MergeCommand())
	root.AddCommand(CountCommand())
	root.AddCommand(RunFnCommand())
	root.AddCommand(&cobra.Command{Use: "docs-merge", Long: merge2.Help})
	root.AddCommand(&cobra.Command{Use: "docs-merge3", Long: merge3.Help})
	return root
}
