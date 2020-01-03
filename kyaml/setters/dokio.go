// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &PerformSetters{}

// PerformSetters sets field values
type PerformSetters struct {
	// Name is the name of the setter to perform
	Name string

	// Value is the value to set
	Value string

	// Description, if set will annotate the field with a description.
	Description string

	// SetBy, if set will annotate the field with who set it.
	SetBy string

	// Count is set by Filter and is the number of fields modified.
	Count int
}

func (s *PerformSetters) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range input {
		p := &fieldSetter{
			Name:        s.Name,
			Value:       s.Value,
			Description: s.Description,
			SetBy:       s.SetBy,
		}
		if err := input[i].PipeE(p); err != nil {
			return nil, err
		}
		s.Count += p.Count
	}
	return input, nil
}
