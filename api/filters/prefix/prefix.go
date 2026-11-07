// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package prefix

import (
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/internal/affix"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter applies resource name prefix's using the fieldSpecs
type Filter struct {
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`

	FieldSpec types.FieldSpec `json:"fieldSpec,omitempty" yaml:"fieldSpec,omitempty"`

	trackableSetter filtersutil.TrackableSetter
}

var _ kio.Filter = Filter{}
var _ kio.TrackableFilter = &Filter{}

// WithMutationTracker registers a callback which will be invoked each time a field is mutated
func (f *Filter) WithMutationTracker(callback func(key, value, tag string, node *yaml.RNode)) {
	f.trackableSetter.WithMutationTracker(callback)
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return affix.Apply(nodes, f.FieldSpec, f.Prefix, "", f.trackableSetter) //nolint:wrapcheck // Preserve the filter's existing error chain.
}
