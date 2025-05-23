// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// Patch represent either a Strategic Merge Patch or a JSON patch
// and its targets.
// The content of the patch can either be from a file
// or from an inline string.
type PatchArgs struct {
	// AllowNameChange allows name changes to the resource.
	AllowNameChange bool `json:"allowNameChange,omitempty" yaml:"allowNameChange,omitempty"`

	// AllowKindChange allows kind changes to the resource.
	AllowKindChange bool `json:"allowKindChange,omitempty" yaml:"allowKindChange,omitempty"`

	// AllowNoTargetMatch allows files rendering in case of no target (`api/types/selector`) match.
	AllowNoTargetMatch bool `json:"allowNoTargetMatch,omitempty" yaml:"allowNoTargetMatch,omitempty"`
}
