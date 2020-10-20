// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The kustomize CLI.
package main

import (
	"os"

	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands"

	// initialize auth
	// This is here rather than in the libraries because of
	// https://github.com/kubernetes-sigs/kustomize/issues/2060
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
