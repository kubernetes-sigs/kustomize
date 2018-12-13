/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package edit

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/pkg/commands/edit/add"
	"sigs.k8s.io/kustomize/pkg/commands/edit/fix"
	"sigs.k8s.io/kustomize/pkg/commands/edit/set"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

// NewCmdEdit returns an instance of 'edit' subcommand.
func NewCmdEdit(fsys fs.FileSystem, v ifc.Validator, kf ifc.KunstructuredFactory) *cobra.Command {
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
		add.NewCmdAdd(fsys, v, kf),
		set.NewCmdSet(fsys, v),
		fix.NewCmdFix(fsys),
	)
	return c
}
