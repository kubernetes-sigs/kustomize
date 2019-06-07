// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run sigs.k8s.io/kustomize/plugin/pluginator
package main

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sort"
)

// Sort the resources using an ordering defined in the Gvk class.
// This puts cluster-wide basic resources with no
// dependencies (like Namespace, StorageClass, etc.)
// first, and resources with a high number of dependencies
// (like ValidatingWebhookConfiguration) last.
type plugin struct {}

var KustomizePlugin plugin

// Nothing needed for configuration.
func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	resources := make([]*resource.Resource, m.Size())
	ids := m.AllIds()
	sort.Sort(resmap.IdSlice(ids))
	for i, id := range ids {
		resources[i] = m.GetById(id)
	}
	m.Clear()
	for i, id := range ids {
		m.AppendWithId(id, resources[i])
	}
	return nil
}
