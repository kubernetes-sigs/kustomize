// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run sigs.k8s.io/kustomize/v3/cmd/pluginator
package main

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/yaml"
	"sort"
)

// Sort the resmap using an ordering defined in the KindOrder parameter.
// This plugin is a mix of the kustomize legacyordertransformer.go and
// the helm kinder_sorter.go
type plugin struct {
	KindOrder []string `json:"kindorder,omitempty" yaml:"kindorder,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// Nothing needed for configuration.
func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	p.KindOrder = []string{}
	return yaml.Unmarshal(c, p)
}

//
func (p *plugin) GetSortOrder() []string {
	if len(p.KindOrder) != 0 {
		return p.KindOrder
	}

	return []string{
		"APIService",
		"Ingress",
		"Service",
		"CronJob",
		"Job",
		"StatefulSet",
		"Deployment",
		"ReplicaSet",
		"ReplicationController",
		"Pod",
		"DaemonSet",
		"RoleBinding",
		"Role",
		"ClusterRoleBinding",
		"ClusterRole",
		"CustomResourceDefinition",
		"ServiceAccount",
		"PersistentVolumeClaim",
		"PersistentVolume",
		"StorageClass",
		"ConfigMap",
		"Secret",
		"PodSecurityPolicy",
		"LimitRange",
		"ResourceQuota",
		"Namespace",
	}
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	resources := make([]*resource.Resource, m.Size())

	ids := m.AllIds()
	ks := newKubectlDeleteSorter(ids, p.GetSortOrder())
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

type kubectldeleteSorter struct {
	ordering map[string]int
	resids   []resid.ResId
}

func newKubectlDeleteSorter(m []resid.ResId, s []string) *kubectldeleteSorter {
	o := make(map[string]int, len(s))
	for v, k := range s {
		o[k] = v + 1
	}

	return &kubectldeleteSorter{
		resids:   m,
		ordering: o,
	}
}

func (k *kubectldeleteSorter) Len() int { return len(k.resids) }

func (k *kubectldeleteSorter) Swap(i, j int) { k.resids[i], k.resids[j] = k.resids[j], k.resids[i] }

func (k *kubectldeleteSorter) Less(i, j int) bool {
	a := k.resids[i]
	b := k.resids[j]
	first, aok := k.ordering[a.Kind]
	second, bok := k.ordering[b.Kind]
	// if same kind (including unknown) sub sort alphanumeric
	if first == second {
		// if both are unknown and of different kind sort by kind alphabetically
		if !aok && !bok && a.Kind != b.Kind {
			return a.Kind > b.Kind
		}
		return a.String() > b.String()
	}
	// unknown kind is first
	if !aok {
		return true
	}
	if !bok {
		return false
	}
	// sort different kinds
	return first < second
}
