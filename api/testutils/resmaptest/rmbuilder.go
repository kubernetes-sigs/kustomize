// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmaptest_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

// Builds ResMaps for tests, with test-aware error handling.
type rmBuilder struct {
	t  *testing.T
	m  resmap.ResMap
	rf *resource.Factory
}

func NewSeededRmBuilder(t *testing.T, rf *resource.Factory, m resmap.ResMap) *rmBuilder {
	t.Helper()
	return &rmBuilder{t: t, rf: rf, m: m}
}

func NewRmBuilder(t *testing.T, rf *resource.Factory) *rmBuilder {
	t.Helper()
	return NewSeededRmBuilder(t, rf, resmap.New())
}

func NewRmBuilderDefault(t *testing.T) *rmBuilder {
	t.Helper()
	return NewSeededRmBuilderDefault(t, resmap.New())
}

func NewSeededRmBuilderDefault(t *testing.T, m resmap.ResMap) *rmBuilder {
	t.Helper()
	return NewSeededRmBuilder(
		t, provider.NewDefaultDepProvider().GetResourceFactory(), m)
}

func (rm *rmBuilder) Add(m map[string]interface{}) *rmBuilder {
	return rm.AddR(rm.rf.FromMap(m))
}

func (rm *rmBuilder) AddR(r *resource.Resource) *rmBuilder {
	err := rm.m.Append(r)
	if err != nil {
		rm.t.Fatalf("test setup failure: %v", err)
	}
	return rm
}

func (rm *rmBuilder) AddWithName(n string, m map[string]interface{}) *rmBuilder {
	err := rm.m.Append(rm.rf.FromMapWithNamespaceAndName(resid.DefaultNamespace, n, m))
	if err != nil {
		rm.t.Fatalf("test setup failure: %v", err)
	}
	return rm
}

func (rm *rmBuilder) AddWithNsAndName(ns string, n string, m map[string]interface{}) *rmBuilder {
	err := rm.m.Append(rm.rf.FromMapWithNamespaceAndName(ns, n, m))
	if err != nil {
		rm.t.Fatalf("test setup failure: %v", err)
	}
	return rm
}

func (rm *rmBuilder) ReplaceResource(m map[string]interface{}) *rmBuilder {
	r := rm.rf.FromMap(m)
	_, err := rm.m.Replace(r)
	if err != nil {
		rm.t.Fatalf("test setup failure: %v", err)
	}
	return rm
}

func (rm *rmBuilder) ResMap() resmap.ResMap {
	return rm.m
}
