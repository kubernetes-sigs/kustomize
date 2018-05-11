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

package commands

import (
	"flag"
	"io"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/cmd/kustomize/version"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

// NewDefaultCommand returns the default (aka root) command for kustomize command.
func NewDefaultCommand() *cobra.Command {
	fsys := fs.MakeRealFS()
	stdOut, stdErr := os.Stdout, os.Stderr

	c := &cobra.Command{
		Use:   "kustomize",
		Short: "kustomize manages declarative configuration of Kubernetes",
		Long: `
kustomize manages declarative configuration of Kubernetes.

More info at https://github.com/kubernetes/kubectl/tree/master/cmd/kustomize
`,
	}

	c.AddCommand(
		newCmdBuild(stdOut, stdErr, fsys),
		newCmdDiff(stdOut, stdErr, fsys),
		newCmdEdit(stdOut, stdErr, fsys),
		version.NewCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}

// newCmdEdit returns an instance of 'edit' subcommand.
func newCmdEdit(stdOut, stdErr io.Writer, fsys fs.FileSystem) *cobra.Command {
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
		newCmdAdd(stdOut, stdErr, fsys),
		newCmdSet(stdOut, stdErr, fsys),
	)
	return c
}

// newAddCommand returns an instance of 'add' subcommand.
func newCmdAdd(stdOut, stdErr io.Writer, fsys fs.FileSystem) *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Adds configmap/resource/secret to the kustomization file.",
		Long:  "",
		Example: `
	# Adds a configmap to the kustomization file
	kustomize edit add configmap NAME --from-literal=k=v

	# Adds a secret to the kustomization file
	kustomize edit add secret NAME --from-literal=k=v

	# Adds a resource to the kustomization
	kustomize edit add resource <filepath>

	# Adds a patch to the kustomization
	kustomize edit add patch <filepath>
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdAddResource(stdOut, stdErr, fsys),
		newCmdAddPatch(stdOut, stdErr, fsys),
		newCmdAddConfigMap(stdErr, fsys),
	)
	return c
}

// newSetCommand returns an instance of 'set' subcommand.
func newCmdSet(stdOut, stdErr io.Writer, fsys fs.FileSystem) *cobra.Command {
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
		newCmdSetNamePrefix(stdOut, stdErr, fsys),
	)
	return c
}
