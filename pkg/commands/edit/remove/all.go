// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
)

// NewCmdRemove returns an instance of 'remove' subcommand.
func NewCmdRemove(
	fsys fs.FileSystem,
	ldr ifc.Loader) *cobra.Command {
	c := &cobra.Command{
		Use:   "remove",
		Short: "Removes items from the kustomization file.",
		Long:  "",
		Example: `
	# Removes resources from the kustomization file
	kustomize edit remove resource {filepath} {filepath}
	kustomize edit remove resource {pattern}

	# Removes one or more patches from the kustomization file
	kustomize edit remove patch <filepath>

	# Removes one or more commonLabels from the kustomization file
	kustomize edit remove label {labelKey1},{labelKey2}

	# Removes one or more commonAnnotations from the kustomization file
	kustomize edit remove annotation {annotationKey1},{annotationKey2}
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdRemoveResource(fsys),
		newCmdRemoveLabel(fsys, ldr.Validator().MakeLabelNameValidator()),
		newCmdRemoveAnnotation(fsys, ldr.Validator().MakeAnnotationNameValidator()),
		newCmdRemovePatch(fsys),
	)
	return c
}
