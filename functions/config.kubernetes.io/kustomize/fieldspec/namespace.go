// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/util"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type KustomizeNamespaceFilter struct {
	// commonNamespace is set on metadata.namespace for all Resources
	KustomizeNamespace string `yaml:"commonNamespace,omitempty"`

	kustomizeName string
}

var _ kio.Filter = KustomizeNamespaceFilter{}

func (ns KustomizeNamespaceFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	f := ns.namespaceFilter(ns.KustomizeNamespace)
	_, err := kio.FilterAll(f).Filter(input)
	return input, err
}

func (ns *KustomizeNamespaceFilter) SetKustomizeName(n string) *KustomizeNamespaceFilter {
	ns.kustomizeName = n
	return ns
}

var namespaceFieldSpecList *FieldSpecList

func getNamespaceReferenceFieldSpecs() *FieldSpecList {
	if namespaceFieldSpecList != nil {
		return namespaceFieldSpecList
	}
	namespaceFieldSpecList = &FieldSpecList{}
	err := yaml.Unmarshal([]byte(namespaceFieldSpecs), &namespaceFieldSpecList)
	if err != nil {
		panic(err)
	}
	return namespaceFieldSpecList
}

func (ns KustomizeNamespaceFilter) namespaceFilter(value string) *FieldSpecListFilter {
	return &FieldSpecListFilter{
		FieldSpecList: *getNamespaceReferenceFieldSpecs(),
		SetValue: func(node *yaml.RNode) error {
			v := yaml.NewScalarRNode(value)
			if err := util.SetSetter(v, ns.kustomizeName); err != nil {
				return err
			}
			return node.PipeE(yaml.FieldSetter{Value: v})
		},
		CreateKind: yaml.ScalarNode,
	}
}

const namespaceFieldSpecs = `
items:
- path: metadata/namespace
  create: true
`
