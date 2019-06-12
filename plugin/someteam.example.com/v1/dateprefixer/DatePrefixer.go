// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

type plugin struct{}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
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
	tr, err := transformers.NewPrefixSuffixTransformer(
		getDate()+"-", "",
		config.MakeDefaultConfig().NamePrefix)
	if err != nil {
		return err
	}
	return tr.Transform(m)
}
