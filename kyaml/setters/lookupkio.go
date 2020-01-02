// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters

import (
	"sort"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &LookupSetters{}

// LookupSetters identifies setters for a collection of Resources
type LookupSetters struct {
	// Name is the name of the setter to match.  Optional.
	Name string

	// SetterCounts is populated by Filter and contains the count of fields matching each setter.
	SetterCounts []setterCount
}

// setterCount records the identified setters and number of fields matching those setters
type setterCount struct {
	// Count is the number of substitutions possible to perform
	Count int

	// setter is the substitution found
	setter
}

// Filter implements kio.Filter
func (l *LookupSetters) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	setters := map[string]*setterCount{}

	for i := range input {
		// lookup substitutions for this object
		ls := &lookupSetters{Name: l.Name}
		if err := input[i].PipeE(ls); err != nil {
			return nil, err
		}

		// aggregate counts for each setter by name.  takes the description and value from
		// the first setter for each name encountered.
		for j := range ls.Setters {
			setter := ls.Setters[j]
			curr, found := setters[setter.Name]
			if !found {
				curr = &setterCount{setter: setter}
				setters[setter.Name] = curr
			}
			curr.Count++
		}
	}

	// pull out and sort the results by setter name
	for _, v := range setters {
		l.SetterCounts = append(l.SetterCounts, *v)
	}
	sort.Slice(l.SetterCounts, func(i, j int) bool {
		return l.SetterCounts[i].Name < l.SetterCounts[j].Name
	})
	return input, nil
}
