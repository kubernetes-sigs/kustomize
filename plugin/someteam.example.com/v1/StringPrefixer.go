// +build plugin

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	Metadata metaData `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type metaData struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	tr, err := transformers.NewNamePrefixSuffixTransformer(
		p.Metadata.Name + "-", "",
		config.MakeDefaultConfig().NamePrefix)
	if err != nil {
		return err
	}
	return tr.Transform(m)
}
