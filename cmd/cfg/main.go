// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"sigs.k8s.io/kustomize/cmd/cfg/cmd"
)

func main() {
	if err := cmd.GetRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
