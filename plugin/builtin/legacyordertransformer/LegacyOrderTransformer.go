// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"sort"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/yaml"
)

// Sort the resources using a customizable ordering based of Kind.
// Defaults to the ordering of the GVK struct, which puts cluster-wide basic
// resources with no dependencies (like Namespace, StorageClass, etc.) first,
// and resources with a high number of dependencies
// (like ValidatingWebhookConfiguration) last.
type plugin struct {
	LegacySortOptions *types.LegacySortOptions `json:"legacySortOptions,omitempty" yaml:"legacySortOptions,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// Nothing needed for configuration.
func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return
	}
	if p.LegacySortOptions == nil {
		p.LegacySortOptions = &types.LegacySortOptions{
			OrderFirst: resid.OrderFirst,
			OrderLast:  resid.OrderLast,
		}
	}
	return
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	resources := make([]*resource.Resource, m.Size())
	ids := m.AllIds()
	s := &idSorter{resIds: ids, sortOptions: p.LegacySortOptions}
	sort.Sort(s)
	for i, id := range s.resIds {
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

// idSorter implements the sort interface.
type idSorter struct {
	resIds      []resid.ResId
	sortOptions *types.LegacySortOptions
}

var _ sort.Interface = idSorter{}

func (a idSorter) Len() int      { return len(a.resIds) }
func (a idSorter) Swap(i, j int) { a.resIds[i], a.resIds[j] = a.resIds[j], a.resIds[i] }
func (a idSorter) Less(i, j int) bool {
	if !a.resIds[i].Gvk.Equals(a.resIds[j].Gvk) {
		return gvkLessThan(a.resIds[i].Gvk, a.resIds[j].Gvk,
			a.sortOptions.OrderFirst, a.sortOptions.OrderLast)
	}
	return a.resIds[i].String() < a.resIds[j].String()
}

func gvkLessThan(gvk1, gvk2 resid.Gvk, orderFirst, orderLast []string) bool {
	var typeOrders = func() map[string]int {
		m := map[string]int{}
		for i, n := range orderFirst {
			m[n] = -len(orderFirst) + i
		}
		for i, n := range orderLast {
			m[n] = 1 + i
		}
		return m
	}()

	index1 := typeOrders[gvk1.Kind]
	index2 := typeOrders[gvk2.Kind]
	if index1 != index2 {
		return index1 < index2
	}
	return gvk1.String() < gvk2.String()
}
