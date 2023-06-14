package main


import (
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Resources        []string `json:"resources,omitempty" yaml:"resources,omitempty" protobuf:"bytes,2,opt,name=resources"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.Resources = []string{}
	err = yaml.Unmarshal(config, p)
	p.h = h
	return
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	var resList []resmap.ResMap

	for _, resourcePath := range p.Resources {
		resources, err := p.h.ResmapFactory().FromFiles(resourcePath)
		if err != nil {
			return nil, err
		}
		resList = append(resList, resources)
	}

	mergedResMap := resmap.ResMap
	for _, resMap := range resList {
		err := mergedResMap.Append(resMap)
		if err != nil {
			return nil, err
			}
	}

	return mergedResMap, nil
}
