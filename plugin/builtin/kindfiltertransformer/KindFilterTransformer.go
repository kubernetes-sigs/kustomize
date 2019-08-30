// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"errors"

	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	// Excluded contains the list of resource names to filter out
	Includes []string `json:"includes,omitempty" yaml:"includes,omitempty"`
	Excludes []string `json:"excludes,omitempty" yaml:"excludes,omitempty"`
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ ifc.Loader, _ *resmap.Factory, c []byte) (err error) {
	p.Includes = []string{}
	p.Excludes = []string{}
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return
	}

	if (len(p.Excludes) != 0) && (len(p.Includes) != 0) {
		return errors.New("Expected includes or excludes set to be specified")
	}

	return err
}

func (p *plugin) Transform(m resmap.ResMap) error {
	ids := m.AllIds()

	if len(p.Excludes) != 0 {
		excludedSet := newResIdSet(p.Excludes)
		for _, id := range ids {
			if excludedSet.In(id) {
				m.Remove(id)
			}
		}
		return nil
	}

	if len(p.Includes) != 0 {
		includedSet := newResIdSet(p.Includes)
		for _, id := range ids {
			if includedSet.NotIn(id) {
				m.Remove(id)
			}
		}
		return nil
	}
	return errors.New("Expected at least one includes or one excludes set to be specified")
}

type ResIdSet struct {
	resids map[string]struct{}
}

func newResIdSet(m []string) *ResIdSet {
	fs := &ResIdSet{
		resids: map[string]struct{}{},
	}
	for _, res := range m {
		fs.resids[res] = struct{}{}
	}
	return fs
}

func (fs *ResIdSet) In(key resid.ResId) bool {
	_, ok := fs.resids[key.Kind]
	return ok
}

func (fs *ResIdSet) NotIn(key resid.ResId) bool {
	_, ok := fs.resids[key.Kind]
	return !ok
}
