// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/kyaml/cmd"
	"sigs.k8s.io/kustomize/cmd/kyaml/docs"
)

//go:generate go run ./docs/gomd docs/md/ docs/gen/
var root = &cobra.Command{
	Use:   "kyaml",
	Short: "kyaml reference comand",
	Long: `Description:
  Reference implementation for using the kyaml libraries.
`,
	Example: ``,
}

func main() {
	root.PersistentFlags().BoolVar(&cmd.StackOnError, "stack-trace", false,
		"print a stack-trace on failure")

	cmd.ExitOnError = true
	root.AddCommand(cmd.GrepCommand())
	root.AddCommand(cmd.TreeCommand())
	root.AddCommand(cmd.CatCommand())
	root.AddCommand(cmd.FmtCommand())
	root.AddCommand(cmd.MergeCommand())
	root.AddCommand(cmd.CountCommand())
	root.AddCommand(cmd.RunFnCommand())
	for i := range docs.Docs {
		root.AddCommand(docs.Docs[i])
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
