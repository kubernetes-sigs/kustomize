// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package complete

import (
	"os"
	"strings"

	"github.com/posener/complete/v2"
	"github.com/posener/complete/v2/predict"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/errors"
)

// NewCommand returns a new install-completion command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "install-completion",
		Short:   commands.CompletionShort,
		Long:    commands.CompletionLong,
		PreRunE: preRunE,
		Run:     run,
	}
}

func preRunE(cmd *cobra.Command, args []string) error {
	// install by default
	if os.Getenv("COMP_INSTALL") == "" {
		if err := errors.Wrap(os.Setenv("COMP_INSTALL", "1")); err != nil {
			return err
		}
	}
	return nil
}

func run(cmd *cobra.Command, args []string) {
	// find the root command
	for cmd.Parent() != nil {
		cmd = cmd.Parent()
	}

	// do completion
	Complete(cmd).Complete("kustomize")
}

// Complete returns a completion command for a cobra command
func Complete(cmd *cobra.Command) *complete.Command {
	cc := &complete.Command{
		Flags: map[string]complete.Predictor{},
		Sub:   map[string]*complete.Command{},
	}

	// add completion for each subcommand
	for i := range cmd.Commands() {
		c := cmd.Commands()[i]
		name := strings.Split(c.Use, " ")[0]
		cc.Sub[name] = Complete(c)
	}

	// add completion for each flag
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		cc.Flags[flag.Name] = predict.Nothing
	})
	return cc
}
