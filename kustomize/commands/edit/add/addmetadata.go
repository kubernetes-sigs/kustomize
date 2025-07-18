// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
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

type addMetadataOptions struct {
	force                 bool
	metadata              map[string]string
	mapValidator          func(map[string]string) error
	kind                  kindOfAdd
	labelsWithoutSelector bool
	includeTemplates      bool
}

// newCmdAddAnnotation adds one or more commonAnnotations to the kustomization file.
func newCmdAddAnnotation(fSys filesys.FileSystem, v func(map[string]string) error) *cobra.Command {
	var o addMetadataOptions
	o.kind = annotation
	o.mapValidator = v
	cmd := &cobra.Command{
		Use: "annotation",
		Short: "Adds one or more commonAnnotations to " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		add annotation {annotationKey1:annotationValue1} {annotationKey2:annotationValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.addAnnotations)
		},
	}
	cmd.Flags().BoolVarP(&o.force, "force", "f", false,
		"overwrite commonAnnotation if it already exists",
	)
	return cmd
}

// newCmdAddLabel adds one or more commonLabels to the kustomization file.
func newCmdAddLabel(fSys filesys.FileSystem, v func(map[string]string) error) *cobra.Command {
	var o addMetadataOptions
	o.kind = label
	o.mapValidator = v
	cmd := &cobra.Command{
		Use: "label",
		Short: "Adds one or more commonLabels to " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		add label {labelKey1:labelValue1} {labelKey2:labelValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.addLabels)
		},
	}
	cmd.Flags().BoolVarP(&o.force, "force", "f", false,
		"overwrite commonLabel if it already exists",
	)
	cmd.Flags().BoolVar(&o.labelsWithoutSelector, "without-selector", false,
		"using add labels without selector option",
	)
	cmd.Flags().BoolVar(&o.includeTemplates, "include-templates", false,
		"include labels in templates (requires --without-selector)",
	)
	return cmd
}

func (o *addMetadataOptions) runE(
	args []string, fSys filesys.FileSystem, adder func(*types.Kustomization) error) error {
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
	err = adder(m)
	if err != nil {
		return err
	}
	return kf.Write(m)
}

// validateAndParse validates `add` commands and parses them into o.metadata
func (o *addMetadataOptions) validateAndParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify %s", o.kind)
	}
	if !o.labelsWithoutSelector && o.includeTemplates {
		return fmt.Errorf("--without-selector flag must be specified for --include-templates to work")
	}
	m, err := util.ConvertSliceToMap(args, o.kind.String())
	if err != nil {
		return err
	}
	if err = o.mapValidator(m); err != nil {
		return err
	}
	o.metadata = m
	return nil
}

func (o *addMetadataOptions) addAnnotations(m *types.Kustomization) error {
	if m.CommonAnnotations == nil {
		m.CommonAnnotations = make(map[string]string)
	}
	return o.writeToMap(m.CommonAnnotations, annotation)
}

func (o *addMetadataOptions) addLabels(m *types.Kustomization) error {
	if o.labelsWithoutSelector {
		return o.writeToLabels(m, label)
	}
	if m.CommonLabels == nil {
		m.CommonLabels = make(map[string]string)
	}
	return o.writeToMap(m.CommonLabels, label)
}

func (o *addMetadataOptions) writeToMap(m map[string]string, kind kindOfAdd) error {
	for k, v := range o.metadata {
		if err := o.writeToMapEntry(m, k, v, kind); err != nil {
			return err
		}
	}
	return nil
}

func (o *addMetadataOptions) writeToMapEntry(m map[string]string, k, v string, kind kindOfAdd) error {
	if _, ok := m[k]; ok && !o.force {
		return fmt.Errorf("%s %s already in kustomization file. Use --force to override.", kind, k)
	}
	m[k] = v
	return nil
}

func (o *addMetadataOptions) writeToLabels(m *types.Kustomization, kind kindOfAdd) error {
	lbl := types.Label{
		Pairs:            make(map[string]string),
		IncludeSelectors: false,
		IncludeTemplates: o.includeTemplates,
	}
	for k, v := range o.metadata {
		if i, ok := o.findLabelKeyIndex(m, lbl, k); ok {
			if err := o.writeToMapEntry(m.Labels[i].Pairs, k, v, kind); err != nil {
				return err
			}
			continue
		}
		if i, ok := o.findLabelIndex(m, lbl); ok {
			if err := o.writeToMapEntry(m.Labels[i].Pairs, k, v, kind); err != nil {
				return err
			}
			continue
		}
		if err := o.writeToMap(lbl.Pairs, kind); err != nil {
			return err
		}
		m.Labels = append(m.Labels, lbl)
	}
	return nil
}

func (o *addMetadataOptions) matchLabelSettings(lbl1, lbl2 types.Label) bool {
	return lbl1.IncludeSelectors == lbl2.IncludeSelectors &&
		lbl1.IncludeTemplates == lbl2.IncludeTemplates
}

func (o *addMetadataOptions) findLabelIndex(m *types.Kustomization, lbl types.Label) (int, bool) {
	for i, ml := range m.Labels {
		if o.matchLabelSettings(ml, lbl) {
			return i, true
		}
	}
	return 0, false
}

func (o *addMetadataOptions) findLabelKeyIndex(m *types.Kustomization, lbl types.Label, key string) (int, bool) {
	if i, found := o.findLabelIndex(m, lbl); found {
		if _, ok := m.Labels[i].Pairs[key]; ok {
			return i, true
		}
	}
	return 0, false
}
