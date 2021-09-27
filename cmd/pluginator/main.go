// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// A code generator. See /plugin/doc.go for an explanation.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/pluginator/v2/internal/builtinplugin"
	"sigs.k8s.io/kustomize/cmd/pluginator/v2/internal/krmfunction"
)

func main() {
	cmd := cobra.Command{
		Use:   "pluginator",
		Short: "pluginator is used to convert a kustomize Go plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return builtinplugin.ConvertToBuiltInPlugin()
		},
	}

	cmd.AddCommand(krmfunction.NewKrmFunctionCmd())

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
