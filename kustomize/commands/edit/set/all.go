// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// NewCmdSet returns an instance of 'set' subcommand.
func NewCmdSet(fSys filesys.FileSystem, ldr ifc.KvLoader, v ifc.Validator) *cobra.Command {
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
		newCmdSetNamePrefix(fSys),
		newCmdSetNameSuffix(fSys),
		newCmdSetNamespace(fSys, v),
		newCmdSetImage(fSys),
		newCmdSetReplicas(fSys),
		newCmdSetLabel(fSys, ldr.Validator().MakeLabelValidator()),
		newCmdSetAnnotation(fSys, ldr.Validator().MakeAnnotationValidator()),
	)
	return c
}
