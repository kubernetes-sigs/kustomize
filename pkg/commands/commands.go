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

// Package commands holds the CLI glue mapping textual commands/args to method calls.
package commands

import (
	"flag"
	"os"

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/spf13/cobra"
)

// NewDefaultCommand returns the default (aka root) command for kustomize command.
func NewDefaultCommand() *cobra.Command {
	fsys := fs.MakeRealFS()
	stdOut := os.Stdout

	c := &cobra.Command{
		Use:   "kustomize",
		Short: "kustomize manages declarative configuration of Kubernetes",
		Long: `
kustomize manages declarative configuration of Kubernetes.

See https://github.com/kubernetes-sigs/kustomize
`,
	}

	c.AddCommand(
		// TODO: Make consistent API for newCmd* functions.
		newCmdBuild(stdOut, fsys),
		newCmdEdit(fsys),
		newCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}

// newCmdEdit returns an instance of 'edit' subcommand.
func newCmdEdit(fsys fs.FileSystem) *cobra.Command {
	c := &cobra.Command{
		Use:   "edit",
		Short: "Edits a kustomization file",
		Long:  "",
		Example: `
	# Adds a configmap to the kustomization file
	kustomize edit add configmap NAME --from-literal=k=v

	# Sets the nameprefix field
	kustomize edit set nameprefix <prefix-value>
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdAdd(fsys),
		newCmdSet(fsys),
	)
	return c
}

// newAddCommand returns an instance of 'add' subcommand.
func newCmdAdd(fsys fs.FileSystem) *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Adds configmap/resource/patch/base to the kustomization file.",
		Long:  "",
		Example: `
	# Adds a configmap to the kustomization file
	kustomize edit add configmap NAME --from-literal=k=v

	# Adds a resource to the kustomization
	kustomize edit add resource <filepath>

	# Adds a patch to the kustomization
	kustomize edit add patch <filepath>

	# Adds one or more base directories to the kustomization
	kustomize edit add base <filepath>
	kustomize edit add base <filepath1>,<filepath2>,<filepath3>

	# Adds one or more commonLabels to the kustomization
	kustomize edit add label {labelKey1:labelValue1},{labelKey2:labelValue2}

	# Adds one or more commonAnnotations to the kustomization
	kustomize edit add annotation {annotationKey1:annotationValue1},{annotationKey2:annotationValue2}
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdAddResource(fsys),
		newCmdAddPatch(fsys),
		newCmdAddConfigMap(fsys),
		newCmdAddBase(fsys),
		newCmdAddLabel(fsys),
		newCmdAddAnnotation(fsys),
	)
	return c
}

// newSetCommand returns an instance of 'set' subcommand.
func newCmdSet(fsys fs.FileSystem) *cobra.Command {
	c := &cobra.Command{
		Use:   "set",
		Short: "Sets the value of different fields in kustomization file.",
		Long:  "",
		Example: `
	# Sets the nameprefix field
	kustomize edit set nameprefix <prefix-value>
`,
		Args: cobra.MinimumNArgs(1),
	}

	c.AddCommand(
		newCmdSetNamePrefix(fsys),
		newCmdSetNamespace(fsys),
		newCmdSetImageTag(fsys),
	)
	return c
}
