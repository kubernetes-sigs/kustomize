// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/openapi/fetch"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/openapi/info"
)

// NewCmdOpenAPI makes a new openapi command.
func NewCmdOpenAPI(w io.Writer) *cobra.Command {

	openApiCmd := &cobra.Command{
		Use:     "openapi",
		Short:   "Commands for interacting with the OpenAPI data",
		Example: `kustomize openapi info`,
		Hidden:  true,
	}

	openApiCmd.AddCommand(info.NewCmdInfo(w))
	openApiCmd.AddCommand(fetch.NewCmdFetch(w))
	return openApiCmd
}
