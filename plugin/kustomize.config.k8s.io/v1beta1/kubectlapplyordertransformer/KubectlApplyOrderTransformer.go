// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run sigs.k8s.io/kustomize/v3/cmd/pluginator
package main

import (
	"sort"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
)

// Sort the resmap using an ordering defined in the KindOrder parameter.
// This plugin is a mix of the kustomize legacyordertransformer.go and
// the helm kinder_sorter.go
type plugin struct {
	KindOrder []string `json:"kindorder,omitempty" yaml:"kindorder,omitempty"`
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// Nothing needed for configuration.
func (p *plugin) Config(
	_ ifc.Loader, _ *resmap.Factory, c []byte) (err error) {
	p.KindOrder = []string{}
	return yaml.Unmarshal(c, p)
}

//
func (p *plugin) GetSortOrder() []string {
	if p.KindOrder != nil && len(p.KindOrder) != 0 {
		return p.KindOrder
	}

	return []string{
		"Namespace",
		"ResourceQuota",
		"LimitRange",
		"PodSecurityPolicy",
		"Secret",
		"ConfigMap",
		"StorageClass",
		"PersistentVolume",
		"PersistentVolumeClaim",
		"ServiceAccount",
		"CustomResourceDefinition",
		"ClusterRole",
		"ClusterRoleBinding",
		"Role",
		"RoleBinding",
		"Service",
		"DaemonSet",
		"Pod",
		"ReplicationController",
		"ReplicaSet",
		"Deployment",
		"StatefulSet",
		"Job",
		"CronJob",
		"Ingress",
		"APIService",
	}
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	resources := make([]*resource.Resource, m.Size())

	ids := m.AllIds()
	ks := newKubectlApplySorter(ids, p.GetSortOrder())
	sort.Sort(ks)

	for i, id := range ks.resids {
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

type kubectlapplySorter struct {
	ordering map[string]int
	resids   []resid.ResId
}

func newKubectlApplySorter(m []resid.ResId, s []string) *kubectlapplySorter {
	o := make(map[string]int, len(s))
	for v, k := range s {
		o[k] = v + 1
	}

	return &kubectlapplySorter{
		resids:   m,
		ordering: o,
	}
}

func (k *kubectlapplySorter) Len() int { return len(k.resids) }

func (k *kubectlapplySorter) Swap(i, j int) { k.resids[i], k.resids[j] = k.resids[j], k.resids[i] }

func (k *kubectlapplySorter) Less(i, j int) bool {
	a := k.resids[i]
	b := k.resids[j]
	first, aok := k.ordering[a.Kind]
	second, bok := k.ordering[b.Kind]
	// if same kind (including unknown) sub sort alphanumeric
	if first == second {
		// if both are unknown and of different kind sort by kind alphabetically
		if !aok && !bok && a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.String() < b.String()
	}
	// unknown kind is last
	if !aok {
		return false
	}
	if !bok {
		return true
	}
	// sort different kinds
	return first < second
}
