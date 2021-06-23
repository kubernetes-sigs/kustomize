// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
)

// NewCmdInfo makes a new info command.
func NewCmdInfo(w io.Writer) *cobra.Command {
	infoCmd := cobra.Command{
		Use:     "info",
		Short:   "Prints the `info` field from the kubernetes OpenAPI data",
		Example: `kustomize openapi info`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(w, kubernetesapi.Info)
		},
	}
	return &infoCmd
}
