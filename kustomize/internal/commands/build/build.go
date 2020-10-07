// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Options contain the options for running a build
type Options struct {
	kustomizationPath string
	outputPath        string
	outOrder          reorderOutput
	fnOptions         types.FnPluginLoadingOptions
}

// NewOptions creates a Options object
func NewOptions(p, o string) *Options {
	return &Options{
		kustomizationPath: p,
		outputPath:        o,
	}
}

var examples = `
To generate the resources specified in 'someDir/kustomization.yaml', run

  kustomize build someDir

The default argument to 'build' is '.' (the current working directory).

The argument can be a URL resolving to a directory
with a kustomization.yaml file, e.g.

  kustomize build \
    github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6

The URL should be formulated as described at
https://github.com/hashicorp/go-getter#url-format
`

// NewCmdBuild creates a new build command.
func NewCmdBuild(out io.Writer) *cobra.Command {
	var o Options
	cmd := &cobra.Command{
		Use: "build {path}",
		Short: "Print configuration per contents of " +
			konfig.DefaultKustomizationFileName(),
		Example:      examples,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunBuild(out)
		},
	}

	cmd.Flags().StringVarP(
		&o.outputPath,
		"output", "o", "",
		"If specified, write the build output to this path.")
	cmd.Flags().BoolVar(
		&o.fnOptions.EnableExec, "enable-exec", false, /*do not change!*/
		"enable support for exec functions -- note: exec functions run arbitrary code -- do not use for untrusted configs!!! (Alpha)")
	cmd.Flags().BoolVar(
		&o.fnOptions.EnableStar, "enable-star", false,
		"enable support for starlark functions. (Alpha)")
	cmd.Flags().BoolVar(
		&o.fnOptions.Network, "network", false,
		"enable network access for functions that declare it")
	cmd.Flags().StringVar(
		&o.fnOptions.NetworkName, "network-name", "bridge",
		"the docker network to run the container in")
	cmd.Flags().StringArrayVar(
		&o.fnOptions.Mounts, "mount", []string{},
		"a list of storage options read from the filesystem")
	cmd.Flags().StringArrayVarP(
		&o.fnOptions.Env, "env", "e", []string{},
		"a list of environment variables to be used by functions")

	addFlagLoadRestrictor(cmd.Flags())
	addFlagEnablePlugins(cmd.Flags())
	addFlagReorderOutput(cmd.Flags())
	addFlagEnableManagedbyLabel(cmd.Flags())
	addFlagEnableKyaml(cmd.Flags())

	return cmd
}

// Validate validates build command.
func (o *Options) Validate(args []string) (err error) {
	if len(args) > 1 {
		return errors.New(
			"specify one path to " +
				konfig.DefaultKustomizationFileName())
	}
	if len(args) == 0 {
		o.kustomizationPath = filesys.SelfDir
	} else {
		o.kustomizationPath = args[0]
	}
	err = validateFlagLoadRestrictor()
	if err != nil {
		return err
	}
	o.outOrder, err = validateFlagReorderOutput()
	return
}

func (o *Options) makeOptions() *krusty.Options {
	opts := krusty.MakeDefaultOptions()
	opts.DoLegacyResourceSort = o.outOrder == legacy
	opts.LoadRestrictions = getFlagLoadRestrictorValue()
	if isFlagEnablePluginsSet() {
		c, err := konfig.EnabledPluginConfig(types.BploUseStaticallyLinked)
		if err != nil {
			log.Fatal(err)
		}
		c.FnpLoadingOptions = o.fnOptions
		opts.PluginConfig = c
	}
	opts.AddManagedbyLabel = isManagedbyLabelEnabled()
	opts.UseKyaml = flagEnableKyamlValue
	return opts
}

func (o *Options) RunBuild(out io.Writer) error {
	fSys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(fSys, o.makeOptions())
	m, err := k.Run(o.kustomizationPath)
	if err != nil {
		return err
	}
	return o.emitResources(out, fSys, m)
}

func (o *Options) emitResources(
	out io.Writer, fSys filesys.FileSystem, m resmap.ResMap) error {
	if o.outputPath != "" && fSys.IsDir(o.outputPath) {
		return writeIndividualFiles(fSys, o.outputPath, m)
	}
	res, err := m.AsYaml()
	if err != nil {
		return err
	}
	if o.outputPath != "" {
		return fSys.WriteFile(o.outputPath, res)
	}
	_, err = out.Write(res)
	return err
}

func writeIndividualFiles(
	fSys filesys.FileSystem, folderPath string, m resmap.ResMap) error {
	byNamespace := m.GroupedByCurrentNamespace()
	for namespace, resList := range byNamespace {
		for _, res := range resList {
			fName := fileName(res)
			if len(byNamespace) > 1 {
				fName = strings.ToLower(namespace) + "_" + fName
			}
			err := writeFile(fSys, folderPath, fName, res)
			if err != nil {
				return err
			}
		}
	}
	for _, res := range m.NonNamespaceable() {
		err := writeFile(fSys, folderPath, fileName(res), res)
		if err != nil {
			return err
		}
	}
	return nil
}

func fileName(res *resource.Resource) string {
	return strings.ToLower(res.GetGvk().StringWoEmptyField()) +
		"_" + strings.ToLower(res.GetName()) + ".yaml"
}

func writeFile(
	fSys filesys.FileSystem, path, fName string, res *resource.Resource) error {
	out, err := yaml.Marshal(res.Map())
	if err != nil {
		return err
	}
	return fSys.WriteFile(filepath.Join(path, fName), out)
}
