// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Deprecated: DatePrefixer will be removed with kustomize/api v1.
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

// Add a date prefix to the name.
// A plugin that adapts another plugin.
type plugin struct {
	t resmap.Transformer
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) makePrefixPluginConfig() ([]byte, error) {
	var s struct {
		Prefix     string            `json:"prefix,omitempty" yaml:"prefix,omitempty"`
		FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	}
	s.Prefix = getDate() + "-"
	s.FieldSpecs = []types.FieldSpec{
		{Path: "metadata/name"},
	}
	return yaml.Marshal(s)
}

func (p *plugin) Config(h *resmap.PluginHelpers, _ []byte) error {
	// Ignore the incoming c, compute new config.
	c, err := p.makePrefixPluginConfig()
	if err != nil {
		return errors.Wrapf(
			err, "dateprefixer makeconfig")
	}
	prefixer := builtins.NewPrefixTransformerPlugin()
	err = prefixer.Config(h, c)
	if err != nil {
		return errors.Wrapf(
			err, "prefix configure")
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
	_, err := fmt.Fprintln(os.Stderr, "Deprecated: DatePrefixer will be removed with kustomize/api v1.")
	if err != nil {
		return err
	}
	return p.t.Transform(m)
}
