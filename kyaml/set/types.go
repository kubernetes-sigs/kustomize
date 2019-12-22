// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
)

// FieldSubstitution is a possible field substitution read from a field
type FieldSubstitution struct {
	// Name is the name of the substitution
	Name string

	// Description is a description of the fields current value
	Description string

	// Value is the current substituted value for the field.
	CurrentValue string

	// Type is the type of the substitution
	Type fieldmeta.FieldValueType

	// Marker is the marker used
	Marker string

	// OwnedBy, if set will annotate the field with an owner.
	OwnedBy string
}
