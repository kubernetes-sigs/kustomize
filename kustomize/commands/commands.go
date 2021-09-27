// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package commands holds the CLI glue mapping textual commands/args to method calls.
package commands

import (
	"flag"
	"io"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/cmd/config/completion"
	"sigs.k8s.io/kustomize/cmd/config/configcobra"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/create"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/edit"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/openapi"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/version"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func makeBuildCommand(fSys filesys.FileSystem, w io.Writer) *cobra.Command {
	cmd := build.NewCmdBuild(
		fSys, build.MakeHelp(konfig.ProgramName, "build"), w)
	// Add build flags that don't appear in kubectl.
	build.AddFunctionAlphaEnablementFlags(cmd.Flags())
	return cmd
}

// NewDefaultCommand returns the default (aka root) command for kustomize command.
func NewDefaultCommand() *cobra.Command {
	fSys := filesys.MakeFsOnDisk()
	stdOut := os.Stdout

	c := &cobra.Command{
		Use:   konfig.ProgramName,
		Short: "Manages declarative configuration of Kubernetes",
		Long: `
Manages declarative configuration of Kubernetes.
See https://sigs.k8s.io/kustomize
`,
	}

	pvd := provider.NewDefaultDepProvider()
	c.AddCommand(
		completion.NewCommand(),
		makeBuildCommand(fSys, stdOut),
		edit.NewCmdEdit(
			fSys, pvd.GetFieldValidator(), pvd.GetResourceFactory(), stdOut),
		create.NewCmdCreate(fSys, pvd.GetResourceFactory()),
		version.NewCmdVersion(stdOut),
		openapi.NewCmdOpenAPI(stdOut),
	)
	configcobra.AddCommands(c, konfig.ProgramName)

	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}
