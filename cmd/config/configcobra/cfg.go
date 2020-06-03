// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configcobra

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

func GetCfg(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cfg",
		Short: "Commands for reading and writing configuration.",
	}

	cmd.AddCommand(commands.AnnotateCommand(name))
	cmd.AddCommand(commands.CatCommand(name))
	cmd.AddCommand(commands.CountCommand(name))
	cmd.AddCommand(commands.CreateSetterCommand(name))
	cmd.AddCommand(commands.CreateSubstitutionCommand(name))
	cmd.AddCommand(commands.FmtCommand(name))
	cmd.AddCommand(commands.GrepCommand(name))
	cmd.AddCommand(commands.InitCommand(name))
	cmd.AddCommand(commands.ListSettersCommand(name))
	cmd.AddCommand(commands.MergeCommand(name))
	cmd.AddCommand(commands.Merge3Command(name))
	cmd.AddCommand(commands.SetCommand(name))
	cmd.AddCommand(commands.TreeCommand(name))

	return cmd
}
