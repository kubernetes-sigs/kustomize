// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/iampolicygenerator"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	types.IAMPolicyGeneratorArgs
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.IAMPolicyGeneratorArgs = types.IAMPolicyGeneratorArgs{}
	err = yaml.Unmarshal(config, p)
	return
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	r := resmap.New()
	err := r.ApplyFilter(iampolicygenerator.Filter{
		IAMPolicyGenerator: p.IAMPolicyGeneratorArgs,
	})
	return r, err
}
