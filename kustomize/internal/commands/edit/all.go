// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package edit

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit/add"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit/fix"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit/remove"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit/set"
)

// NewCmdEdit returns an instance of 'edit' subcommand.
func NewCmdEdit(
	fSys filesys.FileSystem, v ifc.Validator, kf ifc.KunstructuredFactory) *cobra.Command {
	c := &cobra.Command{
		Use:   "edit",
		Short: "Edits a kustomization file",
		Long:  "",
		Example: `
	# Adds a configmap to the kustomization file
	kustomize edit add configmap NAME --from-literal=k=v

	# Sets the nameprefix field
	kustomize edit set nameprefix <prefix-value>

	# Sets the namesuffix field
	kustomize edit set namesuffix <suffix-value>
`,
		Args: cobra.MinimumNArgs(1),
	}

	c.AddCommand(
		add.NewCmdAdd(
			fSys,
			kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), v),
			kf),
		set.NewCmdSet(fSys, v),
		fix.NewCmdFix(fSys),
		remove.NewCmdRemove(fSys, v),
	)
	return c
}
