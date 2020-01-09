// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/cmd/resource/status"

	// This is here rather than in the libraries because of
	// https://github.com/kubernetes-sigs/kustomize/issues/2060
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var root = &cobra.Command{
	Use:   "resource",
	Short: "resource reference command",
}

func main() {
	root.AddCommand(status.StatusCommand())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
