// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Add a string prefix to the name.
// A plugin that adapts another plugin.
type plugin struct {
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	t                resmap.Transformer
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) makePrefixSuffixPluginConfig(n string) ([]byte, error) {
	var s struct {
		Prefix     string
		Suffix     string
		FieldSpecs []types.FieldSpec
	}
	s.Prefix = n + "-"
	s.FieldSpecs = []types.FieldSpec{
		{Path: "metadata/name"},
	}
	return yaml.Marshal(s)
}

func (p *plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	err := yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	c, err = p.makePrefixSuffixPluginConfig(p.Name)
	if err != nil {
		return err
	}
	prefixer := builtins.NewPrefixSuffixTransformerPlugin()
	err = prefixer.Config(h, c)
	if err != nil {
		return errors.Wrapf(
			err, "stringprefixer configure")
	}
	p.t = prefixer
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	return p.t.Transform(m)
}
