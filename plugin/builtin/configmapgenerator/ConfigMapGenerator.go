// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.ConfigMapArgs
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.ConfigMapArgs = types.ConfigMapArgs{}
	err = yaml.Unmarshal(config, p)
	if p.ConfigMapArgs.Name == "" {
		p.ConfigMapArgs.Name = p.Name
	}
	if p.ConfigMapArgs.Namespace == "" {
		p.ConfigMapArgs.Namespace = p.Namespace
	}
	p.h = h
	return
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.h.ResmapFactory().FromConfigMapArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), p.ConfigMapArgs)
}
