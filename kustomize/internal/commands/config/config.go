// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/cmd/config/configcobra"
	"sigs.k8s.io/kustomize/kyaml/commandutil"
)

// NewCmdConfig returns an instance of 'config' subcommand.
func NewCmdConfig(fSys filesys.FileSystem) *cobra.Command {
	c := configcobra.NewConfigCommand("kustomize")

	c.AddCommand(
		newCmdSave(fSys),
	)
	return c
}

type saveOptions struct {
	saveDirectory string
}

func newCmdSave(fSys filesys.FileSystem) *cobra.Command {
	var o saveOptions

	c := &cobra.Command{
		Use:   "save",
		Short: "Save default kustomize transformer configurations to a local directory",
		Long:  "",
		Example: `
	# Save the default transformer configurations to a local directory
	save -d ~/.kustomize/config

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}
			err = o.Complete(fSys)
			if err != nil {
				return err
			}
			return o.RunSave(fSys)
		},
		// Alpha version of config is very different than the current version.
		// If alpha is enabled, don't display this command, but keep it for backwards compatibility.
		Hidden: commandutil.GetAlphaEnabled(),
		Deprecated: `The save command is deprecated and will be removed in a future release.

If you require this command file an issue at https://github.com/kubernetes-sigs/kustomize/issues
so we can capture your requirements.
`,
	}
	c.Flags().StringVarP(
		&o.saveDirectory,
		"directory", "d", "",
		"Directory to save the default transformer configurations")

	return c

}

// Validate validates the saveOptions is not empty
func (o *saveOptions) Validate() error {
	if o.saveDirectory == "" {
		return fmt.Errorf("must specify one local directory to save the default transformer configurations")
	}
	return nil
}

// Complete creates the save directory when the directory doesn't exist
func (o *saveOptions) Complete(fSys filesys.FileSystem) error {
	if !fSys.Exists(o.saveDirectory) {
		return fSys.MkdirAll(o.saveDirectory)
	}
	if fSys.IsDir(o.saveDirectory) {
		return nil
	}
	return fmt.Errorf("%s is not a directory", o.saveDirectory)
}

// RunSave saves the default transformer configurations local directory
func (o *saveOptions) RunSave(fSys filesys.FileSystem) error {
	m := builtinpluginconsts.GetDefaultFieldSpecsAsMap()
	for tname, tcfg := range m {
		filename := filepath.Join(o.saveDirectory, tname+".yaml")
		err := fSys.WriteFile(filename, []byte(tcfg))
		if err != nil {
			return err
		}
	}
	return nil
}
