// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

type addBaseOptions struct {
	baseDirectoryPaths string
}

// newCmdAddBase adds the file path of the kustomize base to the kustomization file.
func newCmdAddBase(fsys fs.FileSystem) *cobra.Command {
	var o addBaseOptions

	cmd := &cobra.Command{
		Use:   "base",
		Short: "Adds one or more bases to the kustomization.yaml in current directory",
		Example: `
		add base {filepath1},{filepath2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddBase(fsys)
		},
	}
	return cmd
}

// Validate validates addBase command.
func (o *addBaseOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify a base directory")
	}
	o.baseDirectoryPaths = args[0]
	return nil
}

// Complete completes addBase command.
func (o *addBaseOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddBase runs addBase command (do real work).
func (o *addBaseOptions) RunAddBase(fSys fs.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	// split directory paths
	paths := strings.Split(o.baseDirectoryPaths, ",")
	for _, path := range paths {
		if !fSys.Exists(path) {
			return errors.New(path + " does not exist")
		}
		if kustfile.StringInSlice(path, m.Resources) {
			return fmt.Errorf("base %s already in kustomization file", path)
		}
		m.Resources = append(m.Resources, path)

	}

	return mf.Write(m)
}
