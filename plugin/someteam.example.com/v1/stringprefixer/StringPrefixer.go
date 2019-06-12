// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

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
	tr, err := transformers.NewPrefixSuffixTransformer(
		p.Metadata.Name+"-", "",
		config.MakeDefaultConfig().NamePrefix)
	if err != nil {
		return err
	}
	return tr.Transform(m)
}
