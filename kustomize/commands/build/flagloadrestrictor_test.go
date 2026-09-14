// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"testing"

	"sigs.k8s.io/kustomize/api/types"
)

func TestValidateFlagLoadRestrictor(t *testing.T) {
	defer func() { theFlags.loadRestrictor = "" }()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{value: "", wantErr: false},
		{value: types.LoadRestrictionsRootOnly.String(), wantErr: false},
		{value: types.LoadRestrictionsNone.String(), wantErr: false},
		{value: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		theFlags.loadRestrictor = tt.value
		err := validateFlagLoadRestrictor()
		if tt.wantErr && err == nil {
			t.Errorf("value %q: expected an error, got none", tt.value)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("value %q: unexpected error: %v", tt.value, err)
		}
	}
}

func TestGetFlagLoadRestrictorValue(t *testing.T) {
	defer func() { theFlags.loadRestrictor = "" }()

	tests := []struct {
		value string
		want  types.LoadRestrictions
	}{
		{value: "", want: types.LoadRestrictionsRootOnly},
		{value: types.LoadRestrictionsRootOnly.String(), want: types.LoadRestrictionsRootOnly},
		{value: types.LoadRestrictionsNone.String(), want: types.LoadRestrictionsNone},
		{value: "none", want: types.LoadRestrictionsNone},
		{value: "bogus", want: types.LoadRestrictionsRootOnly},
	}

	for _, tt := range tests {
		theFlags.loadRestrictor = tt.value
		if got := getFlagLoadRestrictorValue(); got != tt.want {
			t.Errorf("value %q: got %v, want %v", tt.value, got, tt.want)
		}
	}
}
