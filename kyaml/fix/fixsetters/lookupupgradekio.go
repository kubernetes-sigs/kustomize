// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fixsetters

import (
	"sort"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &UpgradeV1Setters{}

// UpgradeV1Setters identifies setters for a collection of Resources to upgrade
type UpgradeV1Setters struct {
	// Name is the name of the setter to match.  Optional.
	Name string

	// SetterCounts is populated by Filter and contains the count of fields matching each setter.
	SetterCounts []setterCount

	// Substitutions are groups of partial setters
	Substitutions []substitution
}

// setterCount records the identified setters and number of fields matching those setters
type setterCount struct {
	// Count is the number of substitutions possible to perform
	Count int

	// setter is the substitution found
	setter
}

// Filter implements kio.Filter
func (l *UpgradeV1Setters) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	setters := map[string]*setterCount{}
	substitutions := map[string]*substitution{}

	for i := range input {
		// lookup substitutions for this object
		ls := &upgradeV1Setters{Name: l.Name}
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

		for j := range ls.Substitutions {
			subst := ls.Substitutions[j]
			_, found := substitutions[subst.Name]
			if !found {
				substitutions[subst.Name] = &subst
			}
		}
	}

	// pull out and sort the results by setter name
	for _, v := range setters {
		l.SetterCounts = append(l.SetterCounts, *v)
	}

	for _, subst := range substitutions {
		l.Substitutions = append(l.Substitutions, *subst)
	}

	sort.Slice(l.Substitutions, func(i, j int) bool {
		return l.Substitutions[i].Name < l.Substitutions[j].Name
	})

	sort.Slice(l.SetterCounts, func(i, j int) bool {
		return l.SetterCounts[i].Name < l.SetterCounts[j].Name
	})
	return input, nil
}
