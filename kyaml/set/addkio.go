// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &SetSubstitutionMarker{}

// Sub performs substitutions
type SetSubstitutionMarker struct {
	// Marker is the marker to set
	Marker Marker

	// ResourceMeta defines the Resource to set the marker on
	ResourceMeta yaml.ResourceMeta
}

func (s *SetSubstitutionMarker) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range input {
		m, err := input[i].GetMeta()
		if err != nil {
			return nil, err
		}
		if s.ResourceMeta.Name != "" && m.Name != s.ResourceMeta.Name {
			continue
		}
		if s.ResourceMeta.Kind != "" && m.Kind != s.ResourceMeta.Kind {
			continue
		}
		if err := input[i].PipeE(&s.Marker); err != nil {
			return nil, err
		}
	}
	return input, nil
}
