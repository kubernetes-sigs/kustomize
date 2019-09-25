// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
)

// NewCmdSet returns an instance of 'set' subcommand.
func NewCmdSet(fsys fs.FileSystem, v ifc.Validator) *cobra.Command {
	c := &cobra.Command{
		Use:   "set",
		Short: "Sets the value of different fields in kustomization file.",
		Long:  "",
		Example: `
	# Sets the nameprefix field
	kustomize edit set nameprefix <prefix-value>

	# Sets the namesuffix field
	kustomize edit set namesuffix <suffix-value>
`,
		Args: cobra.MinimumNArgs(1),
	}

	c.AddCommand(
		newCmdSetNamePrefix(fsys),
		newCmdSetNameSuffix(fsys),
		newCmdSetNamespace(fsys, v),
		newCmdSetImage(fsys),
	)
	return c
}
