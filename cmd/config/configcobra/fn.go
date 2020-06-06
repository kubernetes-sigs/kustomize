// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configcobra

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

func GetFn(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fn",
		Short: "Commands for running functions against configuration.",
	}

	cmd.AddCommand(commands.RunCommand(name))
	cmd.AddCommand(commands.SinkCommand(name))
	cmd.AddCommand(commands.SourceCommand(name))
	cmd.AddCommand(commands.WrapCommand())
	cmd.AddCommand(commands.XArgsCommand())

	return cmd
}
