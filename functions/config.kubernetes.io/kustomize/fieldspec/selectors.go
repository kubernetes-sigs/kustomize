// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/util"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// KustomizeSelectorsFilter sets selectors for all Resources in a package.
// Overrides existing selector values iff the keys match.
type KustomizeSelectorsFilter struct {
	// commonSelectors are the selectors to set
	Selectors map[string]*string `yaml:"commonSelectors,omitempty"`

	kustomizeName string
}

func (ns *KustomizeSelectorsFilter) SetKustomizeName(n string) *KustomizeSelectorsFilter {
	ns.kustomizeName = n
	return ns
}

var _ kio.Filter = KustomizeSelectorsFilter{}

func (sf KustomizeSelectorsFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for k, v := range sf.Selectors {
		f := sf.selectorsFilter(k, v)
		_, err := kio.FilterAll(f).Filter(input)
		if err != nil {
			return nil, err
		}
	}
	return input, nil
}

var selectorFieldSpecList *FieldSpecList

func getSelectorReferenceFieldSpecs() *FieldSpecList {
	if selectorFieldSpecList != nil {
		return selectorFieldSpecList
	}
	selectorFieldSpecList = &FieldSpecList{}
	err := yaml.Unmarshal([]byte(commonSelectorFieldSpecs), &selectorFieldSpecList)
	if err != nil {
		panic(err)
	}
	return selectorFieldSpecList
}

func (sf KustomizeSelectorsFilter) selectorsFilter(key string, value *string) *FieldSpecListFilter {
	if value != nil {
		return &FieldSpecListFilter{
			FieldSpecList: *getSelectorReferenceFieldSpecs(),
			SetValue: func(node *yaml.RNode) error {
				v := yaml.NewScalarRNode(*value)
				if err := util.SetSetter(v, sf.kustomizeName); err != nil {
					return err
				}
				return node.PipeE(yaml.FieldSetter{Name: key, Value: v})
			},
			CreateKind: yaml.MappingNode,
		}
	}

	return &FieldSpecListFilter{
		FieldSpecList: *getSelectorReferenceFieldSpecs(),
		SetValue: func(node *yaml.RNode) error {
			_, err := node.Pipe(yaml.FieldClearer{Name: key})
			return err
		},
		CreateKind: yaml.MappingNode,
	}
}

const commonSelectorFieldSpecs = `
items:

# duck-type supported selectors
#
- path: spec/selector/matchLabels
  create: false

# non-duck-type supported selectors
#
- path: spec/selector/matchLabels
  create: true
  group: apps
  kind: StatefulSet
- path: spec/selector/matchLabels
  create: true
  kind: DaemonSet
- path: spec/selector/matchLabels
  create: true
  kind: ReplicaSet
- path: spec/selector/matchLabels
  create: true
  kind: Deployment
- path: spec/selector
  create: true
  version: v1
  kind: Service
- path: spec/selector
  create: true
  version: v1
  kind: ReplicationController
- path: spec/jobTemplate/spec/selector/matchLabels
  create: false
  group: batch
  kind: CronJob
- path: spec/podSelector/matchLabels
  create: false
  group: networking.k8s.io
  kind: NetworkPolicy
- path: spec/ingress/from/podSelector/matchLabels
  create: false
  group: networking.k8s.io
  kind: NetworkPolicy
- path: spec/egress/to/podSelector/matchLabels
  create: false
  group: networking.k8s.io
  kind: NetworkPolicy
`
