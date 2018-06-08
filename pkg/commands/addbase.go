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

package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

type addBaseOptions struct {
	baseDirectoryPath string
}

// newCmdAddBase adds the file path of the kustomize base to the kustomization file.
func newCmdAddBase(fsys fs.FileSystem) *cobra.Command {
	var o addBaseOptions

	cmd := &cobra.Command{
		Use:   "base",
		Short: "Adds a directory path to a base kustomization to the current directory's kustomization file.",
		Example: `
		add base {filepath}`,
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
	o.baseDirectoryPath = args[0]
	return nil
}

// Complete completes addBase command.
func (o *addBaseOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddBase runs addBase command (do real work).
func (o *addBaseOptions) RunAddBase(fsys fs.FileSystem) error {
	_, err := fsys.Stat(o.baseDirectoryPath)
	if err != nil {
		return err
	}

	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}

	m, err := mf.read()
	if err != nil {
		return err
	}

	if stringInSlice(o.baseDirectoryPath, m.Bases) {
		return fmt.Errorf("base %s already in kustomization file", o.baseDirectoryPath)
	}

	m.Bases = append(m.Bases, o.baseDirectoryPath)

	return mf.write(m)
}
