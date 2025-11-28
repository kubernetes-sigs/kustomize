// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	. "sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

var testCases = [][]testCase{scalarTestCases, listTestCases, elementTestCases, mapTestCases}

func TestMerge(t *testing.T) {
	for i := range testCases {
		for j := range testCases[i] {
			tc := testCases[i][j]
			t.Run(tc.description, func(t *testing.T) {
				actual, err := MergeStrings(tc.source, tc.dest, tc.infer, tc.mergeOptions)
				if !assert.NoError(t, err, tc.description) {
					t.FailNow()
				}
				e, err := filters.FormatInput(bytes.NewBufferString(tc.expected))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				estr := strings.TrimSpace(e.String())
				a, err := filters.FormatInput(bytes.NewBufferString(actual))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				astr := strings.TrimSpace(a.String())
				if !assert.Equal(t, estr, astr, tc.description) {
					t.FailNow()
				}
			})
		}
	}
}

type testCase struct {
	description  string
	source       string
	dest         string
	expected     string
	infer        bool
	mergeOptions yaml.MergeOptions
}

// TestMergeWithKindChange tests the AllowKindChange option which allows
// merging nodes of different kinds (e.g., map to list, map to scalar).
// This is needed for Helm values merging where charts may have placeholder
// types that don't match what users want to provide.
// Reference: https://github.com/kubernetes-sigs/kustomize/issues/5766
// Reference: https://github.com/actions/actions-runner-controller/issues/3819
func TestMergeWithKindChange(t *testing.T) {
	testCases := []struct {
		name            string
		source          string
		dest            string
		expected        string
		allowKindChange bool
		expectError     bool
	}{
		{
			// Issue #5766: Chart has empty map {}, user wants to provide a list
			name: "map to list with AllowKindChange=true (issue #5766)",
			source: `
topologySpreadConstraints:
- maxSkew: 1
  topologyKey: topology.kubernetes.io/zone
`,
			dest: `
topologySpreadConstraints: {}
`,
			expected: `
topologySpreadConstraints:
- maxSkew: 1
  topologyKey: topology.kubernetes.io/zone
`,
			allowKindChange: true,
			expectError:     false,
		},
		{
			name: "map to list without AllowKindChange should fail",
			source: `
topologySpreadConstraints:
- maxSkew: 1
`,
			dest: `
topologySpreadConstraints: {}
`,
			allowKindChange: false,
			expectError:     true,
		},
		{
			// Issue #3819: Chart has map structure, user wants to provide a scalar string
			name: "map to scalar with AllowKindChange=true (issue #3819)",
			source: `
githubConfigSecret: my-custom-secret
`,
			dest: `
githubConfigSecret:
  github_token: ""
`,
			expected: `
githubConfigSecret: my-custom-secret
`,
			allowKindChange: true,
			expectError:     false,
		},
		{
			name: "map to scalar without AllowKindChange should fail",
			source: `
githubConfigSecret: my-custom-secret
`,
			dest: `
githubConfigSecret:
  github_token: ""
`,
			allowKindChange: false,
			expectError:     true,
		},
		{
			// Reverse case: list to map
			name: "list to map with AllowKindChange=true",
			source: `
tls:
  termination: edge
  insecureEdgeTerminationPolicy: Redirect
`,
			dest: `
tls: []
`,
			expected: `
tls:
  termination: edge
  insecureEdgeTerminationPolicy: Redirect
`,
			allowKindChange: true,
			expectError:     false,
		},
		{
			// Scalar to map
			name: "scalar to map with AllowKindChange=true",
			source: `
config:
  enabled: true
  value: 42
`,
			dest: `
config: "default"
`,
			expected: `
config:
  enabled: true
  value: 42
`,
			allowKindChange: true,
			expectError:     false,
		},
		{
			// Nested kind change
			name: "nested kind change with AllowKindChange=true",
			source: `
route:
  tls:
    termination: edge
`,
			dest: `
route:
  tls: []
`,
			expected: `
route:
  tls:
    termination: edge
`,
			allowKindChange: true,
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mergeOpts := yaml.MergeOptions{AllowKindChange: tc.allowKindChange}
			actual, err := MergeStrings(tc.source, tc.dest, false, mergeOpts)

			if tc.expectError {
				assert.Error(t, err, "expected error but got none")
				return
			}

			if !assert.NoError(t, err, tc.name) {
				t.FailNow()
			}

			// Parse and format both expected and actual for comparison
			src, err := yaml.Parse(tc.source)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			srcMap, err := src.Map()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			act, err := yaml.Parse(actual)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			actMap, err := act.Map()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// The source values should be present in the result
			for key := range srcMap {
				assert.Contains(t, actMap, key, "result should contain key from source: %s", key)
			}
		})
	}
}
