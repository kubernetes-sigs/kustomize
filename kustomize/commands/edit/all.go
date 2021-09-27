// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package edit

import (
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/edit/add"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/edit/fix"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/edit/listbuiltin"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/edit/remove"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/edit/set"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// NewCmdEdit returns an instance of 'edit' subcommand.
func NewCmdEdit(
	fSys filesys.FileSystem, v ifc.Validator, rf *resource.Factory,
	w io.Writer) *cobra.Command {
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
			rf),
		set.NewCmdSet(
			fSys,
			kv.NewLoader(loader.NewFileLoaderAtCwd(fSys), v),
			v),
		fix.NewCmdFix(fSys, w),
		remove.NewCmdRemove(fSys, v),
		listbuiltin.NewCmdListBuiltinPlugin(),
	)
	return c
}
