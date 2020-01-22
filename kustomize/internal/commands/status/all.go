// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate $GOBIN/mdtogo docs/commands generateddocs/commands --license=none

package status

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/status/cmd"
	"sigs.k8s.io/kustomize/kyaml/commandutil"
)

func NewCmdStatus() *cobra.Command {
	var c = &cobra.Command{
		Use:    "status",
		Short:  "[Alpha] Commands for working with resource status.",
		Hidden: commandutil.GetAlphaEnabled(),
	}

	if !commandutil.GetAlphaEnabled() {
		c.Short = "[Alpha] To enable set KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true"
		c.Long = "[Alpha] To enable set KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true"
		return c
	}

	c.AddCommand(cmd.FetchCommand())
	c.AddCommand(cmd.WaitCommand())
	c.AddCommand(cmd.EventsCommand())
	return c
}
