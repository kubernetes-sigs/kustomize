// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// MergeKeySpec declares a custom merge key for a list field in a specific
// resource type. This is useful for CRDs that don't have registered OpenAPI
// schemas, where strategic merge patch would otherwise replace list contents
// instead of merging them.
type MergeKeySpec struct {
	// Group is the API group of the resource (e.g. "helm.toolkit.fluxcd.io").
	Group string `json:"group,omitempty" yaml:"group,omitempty"`
	// Version is the API version of the resource (optional).
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Kind is the kind of the resource (e.g. "HelmRelease").
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	// Path is the slash-separated path to the list field
	// (e.g. "spec/values/myapp/env").
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Key is the field name to use as the merge key for items in the list.
	Key string `json:"key,omitempty" yaml:"key,omitempty"`
}
