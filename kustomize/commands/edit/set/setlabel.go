// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setLabelOptions struct {
	metadata         map[string]string
	mapValidator     func(map[string]string) error
	includeSelectors bool
}

// newCmdSetLabel sets one or more commonLabels to the kustomization file.
func newCmdSetLabel(fSys filesys.FileSystem, v func(map[string]string) error) *cobra.Command {
	var o setLabelOptions
	o.mapValidator = v
	cmd := &cobra.Command{
		Use: "label",
		Short: "Sets one or more labels or commonLabels in " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		# Set commonLabels (default)
		set label {labelKey1:labelValue1} {labelKey2:labelValue2}

		# Set commonLabels
		set label --include-selectors=true {labelKey1:labelValue1} {labelKey2:labelValue2}

		# Set labels
		set label --include-selectors=false {labelKey1:labelValue1} {labelKey2:labelValue2}
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.setLabels)
		},
	}
	cmd.Flags().BoolVarP(&o.includeSelectors, "include-selectors", "s", true,
		"include label in selectors",
	)
	return cmd
}

func (o *setLabelOptions) runE(
	args []string, fSys filesys.FileSystem, setter func(*types.Kustomization) error) error {
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
	err = setter(m)
	if err != nil {
		return err
	}
	return kf.Write(m)
}

// validateAndParse validates `set` commands and parses them into o.metadata
func (o *setLabelOptions) validateAndParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify label")
	}
	m, err := util.ConvertSliceToMap(args, "label")
	if err != nil {
		return err
	}
	if err = o.mapValidator(m); err != nil {
		return err
	}
	o.metadata = m
	return nil
}

func (o *setLabelOptions) setLabels(m *types.Kustomization) error {
	if o.includeSelectors {
		if m.CommonLabels == nil {
			m.CommonLabels = make(map[string]string)
		}
		return o.writeToMap(m.CommonLabels)
	} else {
		if m.Labels == nil {
			m.Labels = make([]types.Label, 0)
		}
		return o.writeToLabelsSlice(&m.Labels)
	}
}

func (o *setLabelOptions) writeToMap(m map[string]string) error {
	for k, v := range o.metadata {
		m[k] = v
	}
	return nil
}

func (o *setLabelOptions) writeToLabelsSlice(klabels *[]types.Label) error {
	for k, v := range o.metadata {
		isPresent := false
		for _, kl := range *klabels {
			if _, isPresent = kl.Pairs[k]; isPresent {
				kl.Pairs[k] = v
				break
			}
		}
		if !isPresent {
			*klabels = append(*klabels, types.Label{
				Pairs: map[string]string{
					k: v,
				},
				IncludeSelectors: o.includeSelectors,
			})
		}
	}
	return nil
}
