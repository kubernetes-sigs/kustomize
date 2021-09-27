// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/cmd/config/configcobra"
	"sigs.k8s.io/kustomize/kyaml/commandutil"
)

func main() {
	os.Setenv(commandutil.EnableAlphaCommmandsEnvName, "true")
	cmd := configcobra.AddCommands(&cobra.Command{
		Use:   "kubectl-krm",
		Short: "Plugin for configuration based commands.",
		Long: `Plugin for configuration based commands.

Provides commands for working with Kubernetes resource configuration.
`,
	}, "")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
