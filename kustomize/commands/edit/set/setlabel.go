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
	metadata     map[string]string
	mapValidator func(map[string]string) error
}

// newCmdSetLabel sets one or more commonLabels to the kustomization file.
func newCmdSetLabel(fSys filesys.FileSystem, v func(map[string]string) error) *cobra.Command {
	var o setLabelOptions
	o.mapValidator = v
	cmd := &cobra.Command{
		Use: "label",
		Short: "Sets one or more commonLabels in " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		set label {labelKey1:labelValue1} {labelKey2:labelValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.setLabels)
		},
	}
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
	if m.CommonLabels == nil {
		m.CommonLabels = make(map[string]string)
	}
	return o.writeToMap(m.CommonLabels)
}

func (o *setLabelOptions) writeToMap(m map[string]string) error {
	for k, v := range o.metadata {
		m[k] = v
	}
	return nil
}
