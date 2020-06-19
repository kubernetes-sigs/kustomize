// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/filters/annotations"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/yaml"
)

// Add the given annotations to the given field specifications.
type plugin struct {
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	FieldSpecs  []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Annotations = nil
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		err := filtersutil.ApplyToJSON(annotations.Filter{
			Annotations: p.Annotations,
			FsSlice:     p.FieldSpecs,
		}, r)
		if err != nil {
			return err
		}
	}
	return nil
}
