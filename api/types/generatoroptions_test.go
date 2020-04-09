// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"reflect"
	"testing"

	. "sigs.k8s.io/kustomize/api/types"
)

func TestMergeGlobalOptionsIntoLocal(t *testing.T) {
	tests := []struct {
		name     string
		local    *GeneratorOptions
		global   *GeneratorOptions
		expected *GeneratorOptions
	}{
		{
			name:     "everything nil",
			local:    nil,
			global:   nil,
			expected: nil,
		},
		{
			name: "nil global",
			local: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
			},
			global: nil,
			expected: &GeneratorOptions{
				Labels:                map[string]string{"pet": "dog"},
				Annotations:           map[string]string{"fruit": "apple"},
				DisableNameSuffixHash: false,
			},
		},
		{
			name:  "nil local",
			local: nil,
			global: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
			},
			expected: &GeneratorOptions{
				Labels:                map[string]string{"pet": "dog"},
				Annotations:           map[string]string{"fruit": "apple"},
				DisableNameSuffixHash: false,
			},
		},
		{
			name: "global doesn't damage local",
			local: &GeneratorOptions{
				Labels: map[string]string{"pet": "dog"},
				Annotations: map[string]string{
					"fruit": "apple"},
			},
			global: &GeneratorOptions{
				Labels: map[string]string{
					"pet":     "cat",
					"simpson": "homer",
				},
				Annotations: map[string]string{
					"fruit": "peach",
					"tesla": "Y",
				},
			},
			expected: &GeneratorOptions{
				Labels: map[string]string{
					"pet":     "dog",
					"simpson": "homer",
				},
				Annotations: map[string]string{
					"fruit": "apple",
					"tesla": "Y",
				},
				DisableNameSuffixHash: false,
			},
		},
		{
			name: "global disable trumps local",
			local: &GeneratorOptions{
				DisableNameSuffixHash: false,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
		},
		{
			name: "local disable works",
			local: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: false,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
		},
		{
			name: "everyone wants disable",
			local: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
			},
		},
	}
	for _, tc := range tests {
		actual := MergeGlobalOptionsIntoLocal(tc.local, tc.global)
		if !reflect.DeepEqual(tc.expected, actual) {
			t.Fatalf("%s annotations: Expected '%v', got '%v'",
				tc.name, tc.expected, *actual)
		}
	}
}
