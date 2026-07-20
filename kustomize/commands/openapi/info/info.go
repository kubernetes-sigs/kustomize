// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

// NewCmdInfo makes a new info command.
func NewCmdInfo(w io.Writer) *cobra.Command {
	infoCmd := cobra.Command{
		Use:     "info",
		Short:   "Prints the version of the builtin kubernetes OpenAPI data",
		Example: `kustomize openapi info`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(w, "{title:Kubernetes,version:%s}\n", openapi.GetSchemaVersion())
		},
	}
	return &infoCmd
}
