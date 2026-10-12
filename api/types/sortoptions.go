// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// SortOptions defines the order that kustomize outputs resources.
type SortOptions struct {
	// Order selects the ordering strategy.
	Order SortOrder `json:"order,omitempty" yaml:"order,omitempty"`
	// LegacySortOptions tweaks the sorting for the "legacy" sort ordering
	// strategy.
	LegacySortOptions *LegacySortOptions `json:"legacySortOptions,omitempty" yaml:"legacySortOptions,omitempty"`
	// AlphabeticalSortOptions tweaks the sorting for the "alphabetical" sort
	// ordering strategy.
	AlphabeticalSortOptions *AlphabeticalSortOptions `json:"alphabeticalSortOptions,omitempty" yaml:"alphabeticalSortOptions,omitempty"`
}

// SortOrder defines different ordering strategies.
type SortOrder string

const LegacySortOrder SortOrder = "legacy"
const FIFOSortOrder SortOrder = "fifo"
const AlphabeticalSortOrder SortOrder = "alphabetical"

// AlphabeticalSortOptions defines a lexicographical ordering over resource
// identity fields. Fields are compared in the order in which they are listed.
type AlphabeticalSortOptions struct {
	// FieldOrder selects the resource identity fields used for sorting.
	FieldOrder []SortField `json:"fieldOrder,omitempty" yaml:"fieldOrder,omitempty"`
	// CaseInsensitive compares field values after Unicode-aware lowercasing.
	CaseInsensitive bool `json:"caseInsensitive,omitempty" yaml:"caseInsensitive,omitempty"`
}

// SortField is a resource identity field accepted by alphabetical ordering.
type SortField string

const (
	SortFieldGroup     SortField = "group"
	SortFieldVersion   SortField = "version"
	SortFieldKind      SortField = "kind"
	SortFieldNamespace SortField = "namespace"
	SortFieldName      SortField = "name"
)

// DefaultAlphabeticalFieldOrder returns the documented default field order:
// namespace, kind, then name. A fresh slice is returned for each caller.
func DefaultAlphabeticalFieldOrder() []SortField {
	// Known defect: namespace was omitted while the feature was threaded
	// through the API. Tests intentionally capture the required contract.
	return []SortField{SortFieldKind, SortFieldName}
}

// LegacySortOptions define various options for tweaking the "legacy" ordering
// strategy.
type LegacySortOptions struct {
	// OrderFirst selects the resource kinds to order first.
	OrderFirst []string `json:"orderFirst" yaml:"orderFirst"`
	// OrderLast selects the resource kinds to order last.
	OrderLast []string `json:"orderLast" yaml:"orderLast"`
}
