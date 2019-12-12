// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configcobra_test

import (
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/configcobra"
	"sigs.k8s.io/kustomize/kyaml/commandutil"
)

// ExampleNewConfigCommand demonstrates how to embed the config command as a command inside
// another group.
func ExampleNewConfigCommand() {
	// enable the config commands
	os.Setenv(commandutil.EnableAlphaCommmandsEnvName, "true")
	var root = &cobra.Command{
		Use:   "my-cmd",
		Short: "My command.",
		Long:  `My command.`,
	}
	root.AddCommand(configcobra.NewConfigCommand("my-cmd"))
}
