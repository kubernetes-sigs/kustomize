// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addGeneratorOptions struct {
	generatorFilePaths []string
}

// newCmdAddGenerator adds the name of a file containing a generator
// configuration to the kustomization file.
func newCmdAddGenerator(fSys filesys.FileSystem) *cobra.Command {
	var o addGeneratorOptions
	cmd := &cobra.Command{
		Use:   "generator",
		Short: "Add the name of a file containing a generator configuration to the kustomization file",
		Example: `
		add generator {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(fSys, args)
			if err != nil {
				return err
			}
			return o.RunAddGenerator(fSys)
		},
	}
	return cmd
}

// Validate validates add generator command.
func (o *addGeneratorOptions) Validate(fSys filesys.FileSystem, args []string) error {
	// TODO: Add validation for the format of the generator.
	if len(args) == 0 {
		return errors.New("must specify a yaml file which contains a generator plugin resource")
	}
	var err error
	o.generatorFilePaths, err = util.GlobPatterns(fSys, args)
	return err
}

// RunAddGenerator runs add generator command (do real work).
func (o *addGeneratorOptions) RunAddGenerator(fSys filesys.FileSystem) error {
	if len(o.generatorFilePaths) == 0 {
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
	for _, t := range o.generatorFilePaths {
		if kustfile.StringInSlice(t, m.Generators) {
			log.Printf("generator %s already in kustomization file", t)
			continue
		}
		m.Generators = append(m.Generators, t)
	}
	return mf.Write(m)
}
