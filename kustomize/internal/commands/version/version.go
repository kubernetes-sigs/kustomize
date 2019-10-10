// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/v3/provenance"
)

// NewCmdVersion makes a new version command.
func NewCmdVersion(w io.Writer) *cobra.Command {
	var short bool

	versionCmd := cobra.Command{
		Use:     "version",
		Short:   "Prints the kustomize version",
		Example: `kustomize version`,
		Run: func(cmd *cobra.Command, args []string) {
			provenance.GetProvenance().Print(w, short)
		},
	}

	versionCmd.Flags().BoolVar(&short, "short", false, "short form")
	return &versionCmd
}
