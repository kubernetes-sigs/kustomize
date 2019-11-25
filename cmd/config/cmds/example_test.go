// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmds_test

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/cmds"
)

// ExampleNewConfigCommand demonstrates how to embed the config command as a command inside
// another group.
func ExampleNewConfigCommand() {
	var root = &cobra.Command{
		Use:   "my-cmd",
		Short: "My command.",
		Long:  `My command.`,
	}
	root.AddCommand(cmds.NewConfigCommand("my-cmd"))
}
