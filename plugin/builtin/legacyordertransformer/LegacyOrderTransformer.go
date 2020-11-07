// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sort"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

// Sort the resources using an ordering defined in the Gvk class.
// This puts cluster-wide basic resources with no
// dependencies (like Namespace, StorageClass, etc.)
// first, and resources with a high number of dependencies
// (like ValidatingWebhookConfiguration) last.
type plugin struct{}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// Nothing needed for configuration.
func (p *plugin) Config(_ *resmap.PluginHelpers, _ []byte) (err error) {
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	resources := make([]*resource.Resource, m.Size())
	ids := m.AllIds()
	sort.Sort(resmap.IdSlice(ids))
	for i, id := range ids {
		resources[i], err = m.GetByCurrentId(id)
		if err != nil {
			return errors.Wrap(err, "expected match for sorting")
		}
	}
	m.Clear()
	for _, r := range resources {
		m.Append(r)
	}
	return nil
}
