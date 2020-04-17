package patchstrategicmerge

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

type Filter struct {
	Patch *yaml.RNode
}

var _ kio.Filter = Filter{}

func (pf Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(pf.run)).Filter(nodes)
}

func (pf Filter) run(node *yaml.RNode) (*yaml.RNode, error) {
	return merge2.Merge(pf.Patch, node)
}
