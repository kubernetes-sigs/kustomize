// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package commands holds the CLI glue mapping textual commands/args to method calls.
package commands

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/k8sdeps/validator"
	"sigs.k8s.io/kustomize/v3/pkg/commands/build"
	"sigs.k8s.io/kustomize/v3/pkg/commands/create"
	"sigs.k8s.io/kustomize/v3/pkg/commands/edit"
	"sigs.k8s.io/kustomize/v3/pkg/commands/misc"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
)

// NewDefaultCommand returns the default (aka root) command for kustomize command.
func NewDefaultCommand() *cobra.Command {
	fSys := fs.MakeRealFS()
	stdOut := os.Stdout

	c := &cobra.Command{
		Use:   pgmconfig.ProgramName,
		Short: "Manages declarative configuration of Kubernetes",
		Long: `
Manages declarative configuration of Kubernetes.
See https://sigs.k8s.io/kustomize
`,
	}

	uf := kunstruct.NewKunstructuredFactoryImpl()
	pf := transformer.NewFactoryImpl()
	rf := resmap.NewFactory(resource.NewFactory(uf), pf)
	v := validator.NewKustValidator()
	c.AddCommand(
		build.NewCmdBuild(
			stdOut, fSys, v,
			rf, pf),
		edit.NewCmdEdit(fSys, v, uf),
		create.NewCmdCreate(fSys, uf),
		misc.NewCmdConfig(fSys),
		misc.NewCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}
