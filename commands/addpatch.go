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
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

type addPatchOptions struct {
	patchFilePath string
}

// newCmdAddPatch adds the name of a file containing a patch to the kustomization file.
func newCmdAddPatch(out, errOut io.Writer, fsys fs.FileSystem) *cobra.Command {
	var o addPatchOptions

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Add the name of a file containing a patch to the kustomization file.",
		Example: `
		add patch {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddPatch(out, errOut, fsys)
		},
	}
	return cmd
}

// Validate validates addPatch command.
func (o *addPatchOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify a patch file")
	}
	o.patchFilePath = args[0]
	return nil
}

// Complete completes addPatch command.
func (o *addPatchOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddPatch runs addPatch command (do real work).
func (o *addPatchOptions) RunAddPatch(out, errOut io.Writer, fsys fs.FileSystem) error {
	_, err := fsys.Stat(o.patchFilePath)
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

	if stringInSlice(o.patchFilePath, m.Patches) {
		return fmt.Errorf("patch %s already in kustomization file", o.patchFilePath)
	}

	m.Patches = append(m.Patches, o.patchFilePath)

	return mf.write(m)
}
