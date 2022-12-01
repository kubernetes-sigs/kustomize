// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Deprecated: StringPrefixer will be removed with kustomize/api v1.
package main

import (
	"fmt"
	"os"

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

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) makePrefixPluginConfig(n string) ([]byte, error) {
	var s struct {
		Prefix     string            `json:"prefix,omitempty" yaml:"prefix,omitempty"`
		FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
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
	c, err = p.makePrefixPluginConfig(p.Name)
	if err != nil {
		return err
	}
	prefixer := builtins.NewPrefixTransformerPlugin()
	err = prefixer.Config(h, c)
	if err != nil {
		return errors.Wrapf(
			err, "stringprefixer configure")
	}
	p.t = prefixer
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	_, err := fmt.Fprintln(os.Stderr, "Deprecated: StringPrefixer will be removed with kustomize/api v1.")
	if err != nil {
		return err
	}
	return p.t.Transform(m)
}
