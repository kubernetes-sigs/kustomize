// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"io"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
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

type Help struct {
	Use     string
	Short   string
	Long    string
	Example string
}

func MakeHelp(pgmName, cmdName string) Help {
	fN := konfig.DefaultKustomizationFileName()
	return Help{
		Use:   cmdName + " <dir>",
		Short: "Build a kustomization target from a directory or URL.",
		Long: fmt.Sprintf(`Build a set of KRM resources using a '%s' file.
The <dir> argument must be a path to a directory containing
'%s', or a git repository URL with a path suffix
specifying same with respect to the repository root.
If <dir> is omitted, '.' is assumed.
`, fN, fN),
		Example: fmt.Sprintf(`# Build the current working directory
  %s %s

# Build some shared configuration directory
  %s %s /home/config/production

# Build from github
  %s %s \
     https://github.com/kubernetes-sigs/kustomize.git/examples/helloWorld?ref=v1.0.6
`, pgmName, cmdName, pgmName, cmdName, pgmName, cmdName),
	}
}

// NewCmdBuild creates a new build command.
func NewCmdBuild(help Help, out io.Writer) *cobra.Command {
	var o Options
	cmd := &cobra.Command{
		Use:          help.Use,
		Short:        help.Short,
		Long:         help.Long,
		Example:      help.Example,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(args); err != nil {
				return err
			}
			k := krusty.MakeKustomizer(
				o.ModifyKrustyOptions(krusty.MakeDefaultOptions()),
			)
			fSys := filesys.MakeFsOnDisk()
			m, err := k.Run(fSys, o.kustomizationPath)
			if err != nil {
				return err
			}
			if o.outputPath != "" && fSys.IsDir(o.outputPath) {
				// Ignore io.Writer; write to o.outputPath directly.
				return MakeWriter(fSys).WriteIndividualFiles(o.outputPath, m)
			}
			yml, err := m.AsYaml()
			if err != nil {
				return err
			}
			if o.outputPath != "" {
				// Ignore io.Writer; write to o.outputPath directly.
				return fSys.WriteFile(o.outputPath, yml)
			}
			_, err = out.Write(yml)
			return err
		},
	}

	cmd.Flags().StringVarP(
		&o.outputPath,
		"output", "o", "",
		"If specified, write output to this path.")

	AddFunctionFlags(cmd.Flags(), &o.fnOptions)
	AddFlagLoadRestrictor(cmd.Flags())
	AddFlagEnablePlugins(cmd.Flags())
	AddFlagReorderOutput(cmd.Flags())
	AddFlagEnableManagedbyLabel(cmd.Flags())
	AddFlagAllowResourceIdChanges(cmd.Flags())

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

// ModifyKrustyOptions feeds command line data into the krusty options.
func (o *Options) ModifyKrustyOptions(kOpts *krusty.Options) *krusty.Options {
	kOpts.DoLegacyResourceSort = o.outOrder == legacy
	kOpts.LoadRestrictions = getFlagLoadRestrictorValue()
	if isFlagEnablePluginsSet() {
		c, err := konfig.EnabledPluginConfig(types.BploUseStaticallyLinked)
		if err != nil {
			log.Fatal(err)
		}
		c.FnpLoadingOptions = o.fnOptions
		kOpts.PluginConfig = c
	}
	kOpts.AddManagedbyLabel = isManagedbyLabelEnabled()
	kOpts.UseKyaml = konfig.FlagEnableKyamlDefaultValue
	kOpts.AllowResourceIdChanges = flagAllowResourceIdChangesValue
	return kOpts
}
