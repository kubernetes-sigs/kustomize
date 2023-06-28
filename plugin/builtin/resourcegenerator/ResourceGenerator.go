// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

type plugin struct {
	h        *resmap.PluginHelpers
	resource string
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.h = h
	return
}

func (p *plugin) Generate() (resmap.ResMap, error) {

	resourceBytes := []byte(p.resource) //idiot
	return p.h.ResmapFactory().NewResMapFromBytes(resourceBytes)
}
