// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/api/types"
)

func TestGenArgs_String(t *testing.T) {
	tests := []struct {
		ga       *GenArgs
		expected string
	}{
		{
			ga:       nil,
			expected: "{nilGenArgs}",
		},
		{
			ga:       &GenArgs{},
			expected: "{nsfx:false,beh:unspecified}",
		},
		{
			ga: NewGenArgs(
				&GeneratorArgs{
					Behavior: "merge",
					Options:  &GeneratorOptions{DisableNameSuffixHash: false},
				}),
			expected: "{nsfx:true,beh:merge}",
		},
	}
	for _, test := range tests {
		if test.ga.String() != test.expected {
			t.Fatalf("Expected '%s', got '%s'", test.expected, test.ga.String())
		}
	}
}
