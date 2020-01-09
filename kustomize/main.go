// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The kustomize CLI.
package main

import (
	"os"

	"sigs.k8s.io/kustomize/cmd/config/complete"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands"
	"sigs.k8s.io/kustomize/kustomize/v3/register"
)

func main() {
	register.RegisterCoreKinds()

	cmd := commands.NewDefaultCommand()
	complete.Complete(cmd).Complete("kustomize")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
