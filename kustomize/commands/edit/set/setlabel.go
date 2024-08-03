// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setLabelOptions struct {
	metadata              map[string]string
	mapValidator          func(map[string]string) error
	labelsWithoutSelector bool
	includeTemplates      bool
}

// newCmdSetLabel sets one or more commonLabels to the kustomization file.
func newCmdSetLabel(fSys filesys.FileSystem, v func(map[string]string) error) *cobra.Command {
	var o setLabelOptions
	o.mapValidator = v
	cmd := &cobra.Command{
		Use: "label",
		Short: "Sets one or more commonLabels or labels in " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		set label {labelKey1:labelValue1} {labelKey2:labelValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.setLabels)
		},
	}
	cmd.Flags().BoolVar(&o.labelsWithoutSelector, "without-selector", false,
		"using set labels without selector option",
	)
	cmd.Flags().BoolVar(&o.includeTemplates, "include-templates", false,
		"include labels in templates (requires --without-selector)",
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
	if !o.labelsWithoutSelector && o.includeTemplates {
		return fmt.Errorf("--without-selector flag must be specified for --include-templates to work")
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
	if o.labelsWithoutSelector {
		o.removeDuplicateLabels(m)

		var labelPairs *types.Label
		for _, label := range m.Labels {
			if !label.IncludeSelectors && label.IncludeTemplates == o.includeTemplates {
				labelPairs = &label
				break
			}
		}

		if labelPairs != nil {
			if labelPairs.Pairs == nil {
				labelPairs.Pairs = make(map[string]string)
			}
			return o.writeToMap(labelPairs.Pairs)
		}

		m.Labels = append(m.Labels, types.Label{
			Pairs:            make(map[string]string),
			IncludeSelectors: false,
			IncludeTemplates: o.includeTemplates,
		})
		return o.writeToMap(m.Labels[len(m.Labels)-1].Pairs)
	}

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

// removeDuplicateLabels removes duplicate labels from commonLabels or labels
func (o *setLabelOptions) removeDuplicateLabels(m *types.Kustomization) {
	for k := range o.metadata {
		// delete duplicate label from deprecated common labels
		delete(m.CommonLabels, k)
		for idx, label := range m.Labels {
			// delete label if it's already present in labels with mismatched includeTemplates value
			if label.IncludeTemplates != o.includeTemplates {
				m.Labels = deleteLabel(k, label, m.Labels, idx)
			}
			if label.IncludeSelectors {
				// delete label if it's already present in labels and includes selectors
				m.Labels = deleteLabel(k, label, m.Labels, idx)
			}
		}
	}
}

// deleteLabel deletes label from types.Label
func deleteLabel(key string, label types.Label, labels []types.Label, idx int) []types.Label {
	delete(label.Pairs, key)
	if len(label.Pairs) == 0 {
		// remove empty map label.Pairs from labels
		labels = append(labels[:idx], labels[idx+1:]...)
	}
	return labels
}
