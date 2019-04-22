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

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/k8sdeps/validator"
	"sigs.k8s.io/kustomize/pkg/commands/build"
	"sigs.k8s.io/kustomize/pkg/commands/edit"
	"sigs.k8s.io/kustomize/pkg/commands/misc"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/plugins"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
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

	pluginConfig := plugin.DefaultPluginConfig()

	c.PersistentFlags().BoolVar(
		&pluginConfig.GoEnabled,
		plugin.EnableGoPluginsFlagName,
		false, plugin.EnableGoPluginsFlagHelp)
	// Not advertising this alpha feature.
	c.PersistentFlags().MarkHidden(plugin.EnableGoPluginsFlagName)

	// Configuration for ConfigMap and Secret generators.
	genMetaArgs := types.GeneratorMetaArgs{
		PluginConfig: pluginConfig,
	}
	uf := kunstruct.NewKunstructuredFactoryWithGeneratorArgs(&genMetaArgs)
	rf := resmap.NewFactory(resource.NewFactory(uf))
	c.AddCommand(
		build.NewCmdBuild(
			stdOut, fSys,
			rf,
			transformer.NewFactoryImpl(),
			plugins.NewLoader(pluginConfig, rf)),
		edit.NewCmdEdit(fSys, validator.NewKustValidator(), uf),
		misc.NewCmdConfig(fSys),
		misc.NewCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}
