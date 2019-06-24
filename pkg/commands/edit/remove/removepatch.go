/*
Copyright 2019 The Kubernetes Authors.

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

package remove

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"sigs.k8s.io/kustomize/v3/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/patch"
	"sigs.k8s.io/kustomize/v3/pkg/pgmconfig"
)

type removePatchOptions struct {
	patchFilePaths []string
}

// newCmdRemovePatch removes the name of a file containing a patch from the kustomization file.
func newCmdRemovePatch(fsys fs.FileSystem) *cobra.Command {
	var o removePatchOptions

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Removes one or more patches from " + pgmconfig.KustomizationFileNames[0],
		Example: `
		remove patch {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunRemovePatch(fsys)
		},
	}
	return cmd
}

// Validate validates removePatch command.
func (o *removePatchOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a patch file")
	}
	o.patchFilePaths = args
	return nil
}

// Complete completes removePatch command.
func (o *removePatchOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunRemovePatch runs removePatch command (do real work).
func (o *removePatchOptions) RunRemovePatch(fSys fs.FileSystem) error {
	patches, err := globPatternsFS(fSys, o.patchFilePaths)
	if err != nil {
		return err
	}
	if len(patches) == 0 {
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

	var removePatches []string
	for _, p := range patches {
		if !patch.Exist(m.PatchesStrategicMerge, p) {
			log.Printf("patch %s doesn't exist in kustomization file", p)
			continue
		}
		removePatches = append(removePatches, p)
	}
	m.PatchesStrategicMerge = patch.Delete(m.PatchesStrategicMerge, removePatches...)

	return mf.Write(m)
}
