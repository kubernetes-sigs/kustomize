// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"sort"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &LookupSubstitutions{}

// Sub performs substitutions
type LookupSubstitutions struct {
	// Name is the name of the substitution to match.  If unspecified, all substitutions will
	// be matched.
	Name string

	// SubstitutionCounts are the aggregate substitutions matched.
	SubstitutionCounts []FieldSubstitutionCount
}

type FieldSubstitutionCount struct {
	// Count is the number of substitutions possible to perform
	Count int

	// CountComplete is the number of substitutions that have already been performed
	// independent of this object.
	CountComplete int

	// FieldSubstitution is the substitution found
	FieldSubstitution
}

func (l *LookupSubstitutions) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	subs := map[string]*FieldSubstitutionCount{}
	for i := range input {
		// lookup substitutions for this object
		ls := &lookupSubstitutions{Name: l.Name}
		if err := input[i].PipeE(ls); err != nil {
			return nil, err
		}

		// aggregate counts for each substitution
		for j := range ls.Substitutions {
			sub := ls.Substitutions[j]
			curr, found := subs[sub.Name]
			if !found {
				curr = &FieldSubstitutionCount{FieldSubstitution: sub}
				subs[sub.Name] = curr
			}
			curr.Count++
			if sub.CurrentValue != "" {
				curr.CountComplete++
			}
		}
	}

	// pull out and sort the results
	for _, v := range subs {
		l.SubstitutionCounts = append(l.SubstitutionCounts, *v)
	}
	sort.Slice(l.SubstitutionCounts, func(i, j int) bool {
		return l.SubstitutionCounts[i].Name < l.SubstitutionCounts[j].Name
	})
	return input, nil
}
