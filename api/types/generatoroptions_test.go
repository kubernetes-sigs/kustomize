// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
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
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
			global: nil,
			expected: &GeneratorOptions{
				Labels:                map[string]string{"pet": "dog"},
				Annotations:           map[string]string{"fruit": "apple"},
				DisableNameSuffixHash: false,
				Immutable:             false,
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
		},
		{
			name:  "nil local",
			local: nil,
			global: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
			expected: &GeneratorOptions{
				Labels:                map[string]string{"pet": "dog"},
				Annotations:           map[string]string{"fruit": "apple"},
				DisableNameSuffixHash: false,
				Immutable:             false,
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
		},
		{
			name: "global doesn't damage local",
			local: &GeneratorOptions{
				Labels: map[string]string{"pet": "dog"},
				Annotations: map[string]string{
					"fruit": "apple"},
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
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
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFiles,
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
				Immutable:             false,
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
		},
		{
			name: "global disable trumps local",
			local: &GeneratorOptions{
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "local disable works",
			local: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "everyone wants disable",
			local: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "local FileMerge with unspecified mode takes global mode",
			local: &GeneratorOptions{
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeUnspecified,
				},
			},
			global: &GeneratorOptions{
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
			expected: &GeneratorOptions{
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
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

func TestGeneratorOptions_FileMerge_JSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *GeneratorOptions
		expectError bool
	}{
		{
			name:  "with fileMerge content mode",
			input: `{"fileMerge":{"mode":"content"}}`,
			expected: &GeneratorOptions{
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFileContent,
				},
			},
			expectError: false,
		},
		{
			name:  "with fileMerge files mode",
			input: `{"fileMerge":{"mode":"files"}}`,
			expected: &GeneratorOptions{
				FileMerge: &FileMergeOptions{
					Mode: FileMergeModeFiles,
				},
			},
			expectError: false,
		},
		{
			name:  "without fileMerge",
			input: `{}`,
			expected: &GeneratorOptions{
				FileMerge: nil,
			},
			expectError: false,
		},
		{
			name:        "with invalid mode",
			input:       `{"fileMerge":{"mode":"invalid"}}`,
			expected:    nil,
			expectError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var opts GeneratorOptions
			err := json.Unmarshal([]byte(tc.input), &opts)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, &opts)
			}
		})
	}
}
