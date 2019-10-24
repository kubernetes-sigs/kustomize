// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit/patch"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/util"
)

type addPatchOptions struct {
	patchFilePaths []string
}

// newCmdAddPatch adds the name of a file containing a patch to the kustomization file.
func newCmdAddPatch(fSys filesys.FileSystem) *cobra.Command {
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
			return o.RunAddPatch(fSys)
		},
	}
	return cmd
}

// Validate validates addPatch command.
func (o *addPatchOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a patch file")
	}
	o.patchFilePaths = args
	return nil
}

// Complete completes addPatch command.
func (o *addPatchOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddPatch runs addPatch command (do real work).
func (o *addPatchOptions) RunAddPatch(fSys filesys.FileSystem) error {
	patches, err := util.GlobPatterns(fSys, o.patchFilePaths)
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

	for _, p := range patches {
		if patch.Exist(m.PatchesStrategicMerge, p) {
			log.Printf("patch %s already in kustomization file", p)
			continue
		}
		m.PatchesStrategicMerge = patch.Append(m.PatchesStrategicMerge, p)
	}

	return mf.Write(m)
}
