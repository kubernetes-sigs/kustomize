// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	ldr              ifc.Loader
	rf               *resmap.Factory
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.GeneratorOptions
	types.ConfigMapArgs
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, config []byte) (err error) {
	p.GeneratorOptions = types.GeneratorOptions{}
	p.ConfigMapArgs = types.ConfigMapArgs{}
	err = yaml.Unmarshal(config, p)
	if p.ConfigMapArgs.Name == "" {
		p.ConfigMapArgs.Name = p.Name
	}
	if p.ConfigMapArgs.Namespace == "" {
		p.ConfigMapArgs.Namespace = p.Namespace
	}
	p.ldr = ldr
	p.rf = rf
	return
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.rf.FromConfigMapArgs(p.ldr, &p.GeneratorOptions, p.ConfigMapArgs)
}
