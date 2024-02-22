// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type plugin struct {
	h        *resmap.PluginHelpers
	Resource string `json:"resource" yaml:"resource"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) error {
	p.h = h
	if err := yaml.Unmarshal(config, p); err != nil {
		return fmt.Errorf("failed to unmarshal ResourceGenerator config: %w", err)
	}
	return nil
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	resmap, err := p.h.AccumulateResource(p.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to Accumulate: %w", err)
	}
	return resmap, nil
}
