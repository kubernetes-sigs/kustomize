// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &PerformSubstitutions{}

// Sub performs substitutions
type PerformSubstitutions struct {
	// Name is the name of the substitution to perform
	Name string

	// NewValue is the substitution value
	NewValue string

	// Override if set to true will re-substitute already fields with a new value
	Override bool

	// Revert if set to true will substitute fields back to the marker value
	Revert bool

	// Description, if set will annotate the field with a description.
	Description string

	// OwnedBy, if set will annotate the field with an owner.
	OwnedBy string

	// Count is the number of substitutions performed by Filter.
	Count int
}

func (s *PerformSubstitutions) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range input {
		p := &performSubstitutions{
			Name:        s.Name,
			Override:    s.Override,
			Revert:      s.Revert,
			NewValue:    s.NewValue,
			OwnedBy:     s.OwnedBy,
			Description: s.Description,
		}
		if err := input[i].PipeE(p); err != nil {
			return nil, err
		}
		s.Count += p.Count
	}
	return input, nil
}
