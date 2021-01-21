// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package listbuiltin

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/krusty"
)

// NewCmdListBuiltinPlugin return an instance of list-builtin-plugin
// subcommand
func NewCmdListBuiltinPlugin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha-list-builtin-plugin",
		Short: "[Alpha] List the builtin plugins",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			plugins := krusty.GetBuiltinPluginNames()
			fmt.Print("Builtin plugins:\n\n")
			for _, p := range plugins {
				fmt.Printf(" * %s\n", p)
			}
		},
	}
	return cmd
}
