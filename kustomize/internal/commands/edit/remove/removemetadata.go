// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
)

// kindOfAdd is the kind of metadata being added: label or annotation
type kindOfAdd int

const (
	annotation kindOfAdd = iota
	label
)

func (k kindOfAdd) String() string {
	kinds := [...]string{
		"annotation",
		"label",
	}
	if k < 0 || k > 1 {
		return "Unknown metadatakind"
	}
	return kinds[k]
}

type removeMetadataOptions struct {
	ignore         bool
	metadata       []string
	arrayValidator func([]string) error
	kind           kindOfAdd
}

// newCmdRemoveLabel removes one or more commonAnnotations from the kustomization file.
func newCmdRemoveAnnotation(fSys filesys.FileSystem, v func([]string) error) *cobra.Command {
	var o removeMetadataOptions
	o.kind = label
	o.arrayValidator = v
	cmd := &cobra.Command{
		Use: "annotation",
		Short: "Removes one or more commonAnnotations from " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		remove annotation {annotationKey1},{annotationKey2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.removeAnnotations)
		},
	}
	cmd.Flags().BoolVarP(&o.ignore, "ignore-non-existence", "i", false,
		"ignore error if the given label doesn't exist",
	)
	return cmd
}

// newCmdRemoveLabel removes one or more commonLabels from the kustomization file.
func newCmdRemoveLabel(fSys filesys.FileSystem, v func([]string) error) *cobra.Command {
	var o removeMetadataOptions
	o.kind = label
	o.arrayValidator = v
	cmd := &cobra.Command{
		Use: "label",
		Short: "Removes one or more commonLabels from " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		remove label {labelKey1},{labelKey2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.removeLabels)
		},
	}
	cmd.Flags().BoolVarP(&o.ignore, "ignore-non-existence", "i", false,
		"ignore error if the given label doesn't exist",
	)
	return cmd
}

func (o *removeMetadataOptions) runE(
	args []string, fSys filesys.FileSystem, remover func(*types.Kustomization) error) error {
	err := o.validateAndParse(args)
	if err != nil {
		return err
	}
	kf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := kf.Read()
	if err != nil {
		return err
	}
	err = remover(m)
	if err != nil {
		return err
	}
	return kf.Write(m)
}

// validateAndParse validates `remove` commands and parses them into o.metadata
func (o *removeMetadataOptions) validateAndParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify %s", o.kind)
	}
	if len(args) > 1 {
		return fmt.Errorf("%ss must be comma-separated, with no spaces", o.kind)
	}
	m, err := o.convertToArray(args[0])
	if err != nil {
		return err
	}
	if err = o.arrayValidator(m); err != nil {
		return err
	}
	o.metadata = m
	return nil
}

func (o *removeMetadataOptions) convertToArray(arg string) ([]string, error) {
	inputs := strings.Split(arg, ",")
	result := make([]string, 0, len(inputs))

	for _, input := range inputs {
		if len(input) == 0 {
			return nil, o.makeError(input, "name is empty")
		}
		result = append(result, input)
	}
	return result, nil
}

func (o *removeMetadataOptions) removeAnnotations(m *types.Kustomization) error {
	if m.CommonAnnotations == nil && !o.ignore {
		return fmt.Errorf("commonAnnotations is not defined in kustomization file")
	}
	return o.removeFromMap(m.CommonAnnotations, annotation)
}

func (o *removeMetadataOptions) removeLabels(m *types.Kustomization) error {
	if m.CommonLabels == nil && !o.ignore {
		return fmt.Errorf("commonLabels is not defined in kustomization file")
	}
	return o.removeFromMap(m.CommonLabels, label)
}

func (o *removeMetadataOptions) removeFromMap(m map[string]string, kind kindOfAdd) error {
	for _, k := range o.metadata {
		if _, ok := m[k]; !ok && !o.ignore {
			return fmt.Errorf("%s %s is not defined in kustomization file", kind, k)
		}
		delete(m, k)
	}
	return nil
}

func (o *removeMetadataOptions) makeError(input string, message string) error {
	return fmt.Errorf("invalid %s: '%s' (%s)", o.kind, input, message)
}
