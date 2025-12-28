// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"
	"slices"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addConfigurationOptions struct {
	configurationFilePaths []string
}

// newCmdAddConfiguration adds the name of a file containing a configuration
// to the kustomization file.
func newCmdAddConfiguration(fSys filesys.FileSystem) *cobra.Command {
	var o addConfigurationOptions
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Add the name of a file containing a configuration to the kustomization file",
		Long: `Add the name of a file containing a configuration (e.g., a Kubernetes configuration resource) 
to the kustomization file. Configurations are used to define custom transformer specifications 
for CRDs and other resource types.`,
		Example: `
	# Adds a configuration file to the kustomization
	kustomize edit add configuration <filepath>

	# Adds multiple configuration files
	kustomize edit add configuration <filepath1>,<filepath2>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(fSys, args)
			if err != nil {
				return err
			}
			return o.RunAddConfiguration(fSys)
		},
	}
	return cmd
}

// Validate validates add configuration command.
func (o *addConfigurationOptions) Validate(fSys filesys.FileSystem, args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a yaml file which contains a configuration resource")
	}
	var err error
	o.configurationFilePaths, err = util.GlobPatterns(fSys, args)
	return err
}

// RunAddConfiguration runs add configuration command (do real work).
func (o *addConfigurationOptions) RunAddConfiguration(fSys filesys.FileSystem) error {
	if len(o.configurationFilePaths) == 0 {
		return nil
	}
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	for _, c := range o.configurationFilePaths {
		if slices.Contains(m.Configurations, c) {
			log.Printf("configuration %s already in kustomization file", c)
			continue
		}
		m.Configurations = append(m.Configurations, c)
	}
	return mf.Write(m)
}
