// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"sigs.k8s.io/kustomize/v3/pkg/commands"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
