// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/cmddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/runfn"
)

// GetCatRunner returns a RunFnRunner.
func GetRunFnRunner(name string) *RunFnRunner {
	r := &RunFnRunner{}
	c := &cobra.Command{
		Use:     "run DIR",
		Aliases: []string{"run-fns"},
		Short:   commands.RunFnsShort,
		Long:    commands.RunFnsLong,
		Example: commands.RunFnsExamples,
		RunE:    r.runE,
		Args:    cobra.ExactArgs(1),
	}
	fixDocs(name, c)
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	r.Command = c
	r.Command.Flags().BoolVar(
		&r.DryRun, "dry-run", false, "print results to stdout")
	r.Command.Flags().StringSliceVar(
		&r.FnPaths, "fn-path", []string{},
		"directories containing functions without configuration")
	r.Command.AddCommand(XArgsCommand())
	r.Command.AddCommand(WrapCommand())
	return r
}

func RunFnCommand(name string) *cobra.Command {
	return GetRunFnRunner(name).Command
}

// RunFnRunner contains the run function
type RunFnRunner struct {
	IncludeSubpackages bool
	Command            *cobra.Command
	DryRun             bool
	FnPaths            []string
}

func (r *RunFnRunner) runE(c *cobra.Command, args []string) error {
	rec := runfn.RunFns{Path: args[0], FunctionPaths: r.FnPaths}
	if r.DryRun {
		rec.Output = c.OutOrStdout()
	}
	return rec.Execute()

}
