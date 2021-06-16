// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addPatchOptions struct {
	Patch types.Patch
}

// newCmdAddPatch adds the name of a file containing a patch to the kustomization file.
func newCmdAddPatch(fSys filesys.FileSystem) *cobra.Command {
	var o addPatchOptions
	o.Patch.Target = &types.Selector{}

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Add an item to patches field.",
		Long: `This command will add an item to patches field in the kustomization file.
Each item may:

 - be either a strategic merge patch, or a JSON patch
 - be either a file, or an inline string
 - target a single resource or multiple resources

For more information please see https://kubernetes-sigs.github.io/kustomize/api-reference/kustomization/patches/
`,
		Example: `
		add patch --path {filepath} --group {target group name} --version {target version}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}
			return o.RunAddPatch(fSys)
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

// Validate validates addPatch command.
func (o *addPatchOptions) Validate() error {
	if o.Patch.Patch != "" && o.Patch.Path != "" {
		return errors.New("patch and path can't be set at the same time")
	}
	if o.Patch.Patch == "" && o.Patch.Path == "" {
		return errors.New("must provide either patch or path")
	}
	return nil
}

// RunAddPatch runs addPatch command (do real work).
func (o *addPatchOptions) RunAddPatch(fSys filesys.FileSystem) error {
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
	for _, p := range m.Patches {
		if p.Equals(o.Patch) {
			log.Printf("patch %#v already in kustomization file", p)
			return nil
		}
	}
	m.Patches = append(m.Patches, o.Patch)

	return mf.Write(m)
}
