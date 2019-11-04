// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package commands holds the CLI glue mapping textual commands/args to method calls.
package commands

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/k8sdeps/validator"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/build"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/config"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/create"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/version"
)

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
	uf := kunstruct.NewKunstructuredFactoryImpl()
	v := validator.NewKustValidator()
	c.AddCommand(
		build.NewCmdBuild(stdOut),
		edit.NewCmdEdit(fSys, v, uf),
		create.NewCmdCreate(fSys, uf),
		config.NewCmdConfig(fSys),
		version.NewCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}
