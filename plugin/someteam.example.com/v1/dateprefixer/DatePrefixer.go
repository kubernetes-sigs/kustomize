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

// Add a date prefix to the name.
// A plugin that adapts another plugin.
type plugin struct {
	t resmap.Transformer
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) makePrefixSuffixPluginConfig() ([]byte, error) {
	var s struct {
		Prefix     string
		Suffix     string
		FieldSpecs []types.FieldSpec
	}
	s.Prefix = getDate() + "-"
	s.FieldSpecs = []types.FieldSpec{
		{Path: "metadata/name"},
	}
	return yaml.Marshal(s)
}

func (p *plugin) Config(h *resmap.PluginHelpers, _ []byte) error {
	// Ignore the incoming c, compute new config.
	c, err := p.makePrefixSuffixPluginConfig()
	if err != nil {
		return errors.Wrapf(
			err, "dateprefixer makeconfig")
	}
	prefixer := builtins.NewPrefixSuffixTransformerPlugin()
	err = prefixer.Config(h, c)
	if err != nil {
		return errors.Wrapf(
			err, "prefixsuffix configure")
	}
	p.t = prefixer
	return nil
}

// Returns a constant, rather than
//   time.Now().Format("2006-01-02")
// to make tests happy.
// This is just an example.
func getDate() string {
	return "2018-05-11"
}

func (p *plugin) Transform(m resmap.ResMap) error {
	return p.t.Transform(m)
}
