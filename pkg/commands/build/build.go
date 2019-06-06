/// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/pkg/ifc"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc/transformer"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/plugins"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/target"
	"sigs.k8s.io/yaml"
)

// Options contain the options for running a build
type Options struct {
	kustomizationPath string
	outputPath        string
	loadRestrictor    loader.LoadRestrictorFunc
}

// NewOptions creates a Options object
func NewOptions(p, o string) *Options {
	return &Options{
		kustomizationPath: p,
		outputPath:        o,
		loadRestrictor:    loader.RestrictionRootOnly,
	}
}

var examples = `
Use the file somedir/kustomization.yaml to generate a set of api resources:
    build somedir

Use a url pointing to a remote directory/kustomization.yaml to generate a set of api resources:
    build url
The url should follow hashicorp/go-getter URL format described in
https://github.com/hashicorp/go-getter#url-format

url examples:
  sigs.k8s.io/kustomize//examples/multibases?ref=v1.0.6
  github.com/Liujingfang1/mysql
  github.com/Liujingfang1/kustomize//examples/helloWorld?ref=repoUrl2
`

// NewCmdBuild creates a new build command.
func NewCmdBuild(
	out io.Writer, fSys fs.FileSystem,
	v ifc.Validator, rf *resmap.Factory,
	ptf transformer.Factory) *cobra.Command {
	var o Options

	pluginConfig := plugins.DefaultPluginConfig()
	pl := plugins.NewLoader(pluginConfig, rf)

	cmd := &cobra.Command{
		Use:          "build [path]",
		Short:        "Print current configuration per contents of " + pgmconfig.KustomizationFileNames[0],
		Example:      examples,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunBuild(out, v, fSys, rf, ptf, pl)
		},
	}

	cmd.Flags().StringVarP(
		&o.outputPath,
		"output", "o", "",
		"If specified, write the build output to this path.")
	loader.AddLoadRestrictionsFlag(cmd.Flags())
	plugins.AddEnablePluginsFlag(
		cmd.Flags(), &pluginConfig.Enabled)

	cmd.AddCommand(NewCmdBuildPrune(out, v, fSys, rf, ptf, pl))
	return cmd
}

// Validate validates build command.
func (o *Options) Validate(args []string) (err error) {
	if len(args) > 1 {
		return errors.New(
			"specify one path to " + pgmconfig.KustomizationFileNames[0])
	}
	if len(args) == 0 {
		o.kustomizationPath = loader.CWD
	} else {
		o.kustomizationPath = args[0]
	}
	o.loadRestrictor, err = loader.ValidateLoadRestrictorFlag()
	return
}

// RunBuild runs build command.
func (o *Options) RunBuild(
	out io.Writer, v ifc.Validator, fSys fs.FileSystem,
	rf *resmap.Factory, ptf transformer.Factory,
	pl *plugins.Loader) error {
	ldr, err := loader.NewLoader(
		o.loadRestrictor, v, o.kustomizationPath, fSys)
	if err != nil {
		return err
	}
	defer ldr.Cleanup()
	kt, err := target.NewKustTarget(ldr, rf, ptf, pl)
	if err != nil {
		return err
	}
	m, err := kt.MakeCustomizedResMap()
	if err != nil {
		return err
	}
	return o.emitResources(out, fSys, m)
}

func (o *Options) RunBuildPrune(
	out io.Writer, v ifc.Validator, fSys fs.FileSystem,
	rf *resmap.Factory, ptf transformer.Factory,
	pl *plugins.Loader) error {
	ldr, err := loader.NewLoader(
		o.loadRestrictor, v, o.kustomizationPath, fSys)
	if err != nil {
		return err
	}
	defer ldr.Cleanup()
	kt, err := target.NewKustTarget(ldr, rf, ptf, pl)
	if err != nil {
		return err
	}
	m, err := kt.MakePruneConfigMap()
	if err != nil {
		return err
	}
	return o.emitResources(out, fSys, m)
}

func (o *Options) emitResources(
	out io.Writer, fSys fs.FileSystem, m resmap.ResMap) error {
	if o.outputPath != "" && fSys.IsDir(o.outputPath) {
		return writeIndividualFiles(fSys, o.outputPath, m)
	}
	res, err := m.EncodeAsYaml()
	if err != nil {
		return err
	}
	if o.outputPath != "" {
		return fSys.WriteFile(o.outputPath, res)
	}
	_, err = out.Write(res)
	return err
}

func NewCmdBuildPrune(
	out io.Writer, v ifc.Validator, fSys fs.FileSystem,
	rf *resmap.Factory, ptf transformer.Factory,
	pl *plugins.Loader) *cobra.Command {
	var o Options

	cmd := &cobra.Command{
		Use:          "alpha-inventory [path]",
		Short:        "Print the inventory object which contains a list of all other objects",
		Example:      examples,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunBuildPrune(out, v, fSys, rf, ptf, pl)
		},
	}
	return cmd
}

func writeIndividualFiles(
	fSys fs.FileSystem, folderPath string, m resmap.ResMap) error {
	for _, res := range m.Resources() {
		filename := filepath.Join(
			folderPath,
			fmt.Sprintf(
				"%s_%s.yaml",
				strings.ToLower(res.GetGvk().String()),
				strings.ToLower(res.GetName()),
			),
		)
		out, err := yaml.Marshal(res.Map())
		if err != nil {
			return err
		}
		err = fSys.WriteFile(filename, out)
		if err != nil {
			return err
		}
	}
	return nil
}
