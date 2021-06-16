// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removePatchOptions struct {
	Patch types.Patch
}

// newCmdRemovePatch removes the name of a file containing a patch from the kustomization file.
func newCmdRemovePatch(fSys filesys.FileSystem) *cobra.Command {
	var o removePatchOptions
	o.Patch.Target = &types.Selector{}

	cmd := &cobra.Command{
		Use: "patch",
		Short: "Removes a patch from " +
			konfig.DefaultKustomizationFileName(),
		Long: `Removes a patch from patches field. The fields specified by flags must 
exactly match the patch item to successfully remote the item.`,
		Example: `
		remove patch --path {filepath} --group {target group name} --version {target version}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}
			return o.RunRemovePatch(fSys)
		},
	}
	cmd.Flags().StringVar(&o.Patch.Path, "path", "", "Path to the patch file. Cannot be used with --patch at the same time.")
	cmd.Flags().StringVar(&o.Patch.Patch, "patch", "", "Literal string of patch content. Cannot be used with --path at the same time.")
	cmd.Flags().StringVar(&o.Patch.Target.Group, "group", "", "API group in patch target")
	cmd.Flags().StringVar(&o.Patch.Target.Version, "version", "", "API version in patch target")
	cmd.Flags().StringVar(&o.Patch.Target.Kind, "kind", "", "Resource kind in patch target")
	cmd.Flags().StringVar(&o.Patch.Target.Name, "name", "", "Resource name in patch target")
	cmd.Flags().StringVar(&o.Patch.Target.Namespace, "namespace", "", "Resource namespace in patch target")
	cmd.Flags().StringVar(&o.Patch.Target.AnnotationSelector, "annotation-selector", "", "annotationSelector in patch target")
	cmd.Flags().StringVar(&o.Patch.Target.LabelSelector, "label-selector", "", "labelSelector in patch target")

	return cmd
}

// Validate validates removePatch command.
func (o *removePatchOptions) Validate() error {
	if o.Patch.Patch != "" && o.Patch.Path != "" {
		return errors.New("patch and path can't be set at the same time")
	}
	return nil
}

// RunRemovePatch runs removePatch command (do real work).
func (o *removePatchOptions) RunRemovePatch(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	// Omit target if it's empty
	emptyTarget := types.Selector{}
	if o.Patch.Target != nil && *o.Patch.Target == emptyTarget {
		o.Patch.Target = nil
	}

	var patches []types.Patch
	for _, p := range m.Patches {
		if !p.Equals(o.Patch) {
			patches = append(patches, p)
		}
	}
	if len(patches) == len(m.Patches) {
		log.Printf("patch %s doesn't exist in kustomization file", o.Patch.Patch)
		return nil
	}
	m.Patches = patches

	return mf.Write(m)
}
