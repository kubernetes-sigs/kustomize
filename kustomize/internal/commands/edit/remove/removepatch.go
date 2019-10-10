// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/util"
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
		Short: "Removes one or more patches from " +
			pgmconfig.DefaultKustomizationFileName(),
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
