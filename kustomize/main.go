// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The kustomize CLI.
package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/cmd/complete"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands"
)

func main() {
	cmd := commands.NewDefaultCommand()
	completion(cmd)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// completion performs shell completion if kustomize is being called to provide
// shell completion commands.
func completion(cmd *cobra.Command) {
	// bash shell completion passes the command name as the first argument
	// do this after configuring cmd so it has all the subcommands
	if len(os.Args) > 1 {
		// use the base name in case kustomize is called with an absolute path
		name := filepath.Base(os.Args[1])
		if name == "kustomize" {
			// complete calls kustomize with itself as an argument
			complete.Complete(cmd).Complete("kustomize")
			os.Exit(0)
		}
	}
}
