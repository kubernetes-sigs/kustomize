// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/util"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// KustomizeLabelsFilter sets labels on all Resources in a package, including sub-field
// labels -- e.g. spec.template.metadata.labels.
// Overrides existing labels iff the keys match.
type KustomizeLabelsFilter struct {
	// commonLabels are the labels to set
	Labels map[string]*string `yaml:"commonLabels,omitempty"`

	kustomizeName string
}

func (ns *KustomizeLabelsFilter) SetKustomizeName(n string) *KustomizeLabelsFilter {
	ns.kustomizeName = n
	return ns
}

var _ kio.Filter = KustomizeLabelsFilter{}

func (lf KustomizeLabelsFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for k, v := range lf.Labels {
		f := lf.labelsFilter(k, v)
		_, err := kio.FilterAll(f).Filter(input)
		if err != nil {
			return nil, err
		}
	}
	return input, nil
}

var labelFieldSpecList *FieldSpecList

func getLabelReferenceFieldSpecs() *FieldSpecList {
	if labelFieldSpecList != nil {
		return labelFieldSpecList
	}
	labelFieldSpecList = &FieldSpecList{}
	err := yaml.Unmarshal([]byte(commonLabelFieldSpecs), labelFieldSpecList)
	if err != nil {
		panic(err)
	}
	return labelFieldSpecList
}

func (lf KustomizeLabelsFilter) labelsFilter(key string, value *string) *FieldSpecListFilter {
	if value != nil {
		return &FieldSpecListFilter{
			FieldSpecList: *getLabelReferenceFieldSpecs(),
			SetValue: func(node *yaml.RNode) error {
				v := yaml.NewScalarRNode(*value)
				if err := util.SetSetter(v, lf.kustomizeName); err != nil {
					return err
				}
				return node.PipeE(yaml.FieldSetter{Name: key, Value: v})
			},
			CreateKind: yaml.MappingNode,
		}
	}

	return &FieldSpecListFilter{
		FieldSpecList: *getLabelReferenceFieldSpecs(),
		SetValue: func(node *yaml.RNode) error {
			return node.PipeE(yaml.FieldClearer{Name: key})
		},
		CreateKind: yaml.MappingNode,
	}
}

const commonLabelFieldSpecs = `
items:
# duck-type supported labels
#
- path: metadata/labels
  create: true
- path: spec/template/metadata/labels
  create: false

# non-duck-type supported labels
#
- path: spec/template/metadata/labels
  create: true
  version: v1
  kind: ReplicationController

- path: spec/template/metadata/labels
  create: true
  kind: Deployment

- path: spec/template/metadata/labels
  create: true
  kind: ReplicaSet

- path: spec/template/metadata/labels
  create: true
  kind: DaemonSet

- path: spec/template/metadata/labels
  create: true
  group: apps
  kind: StatefulSet

- path: spec/volumeClaimTemplates[]/metadata/labels
  create: true
  group: apps
  kind: StatefulSet

- path: spec/template/metadata/labels
  create: true
  group: batch
  kind: Job

- path: spec/jobTemplate/metadata/labels
  create: true
  group: batch
  kind: CronJob

- path: spec/jobTemplate/spec/template/metadata/labels
  create: true
  group: batch
  kind: CronJob
`
