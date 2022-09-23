// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addTransformerOptions struct {
	transformerFilePaths []string
}

// newCmdAddTransformer adds the name of a file containing a transformer
// configuration to the kustomization file.
func newCmdAddTransformer(fSys filesys.FileSystem) *cobra.Command {
	var o addTransformerOptions
	cmd := &cobra.Command{
		Use:   "transformer",
		Short: "Add the name of a file containing a transformer configuration to the kustomization file",
		Example: `
		add transformer {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(fSys, args)
			if err != nil {
				return err
			}
			return o.RunAddTransformer(fSys)
		},
	}
	return cmd
}

// Validate validates add transformer command.
func (o *addTransformerOptions) Validate(fSys filesys.FileSystem, args []string) error {
	// TODO: Add validation for the format of the transformer.
	if len(args) == 0 {
		return errors.New("must specify a yaml file which contains a transformer plugin resource")
	}
	var err error
	o.transformerFilePaths, err = util.GlobPatterns(fSys, args)
	return err
}

// RunAddTransformer runs add transformer command (do real work).
func (o *addTransformerOptions) RunAddTransformer(fSys filesys.FileSystem) error {
	if len(o.transformerFilePaths) == 0 {
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
	for _, t := range o.transformerFilePaths {
		if kustfile.StringInSlice(t, m.Transformers) {
			log.Printf("transformer %s already in kustomization file", t)
			continue
		}
		m.Transformers = append(m.Transformers, t)
	}
	return mf.Write(m)
}
