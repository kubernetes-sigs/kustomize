// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestAppendMissingListNode(t *testing.T) {
	tests := []struct {
		name     string
		dst      string
		src      string
		keys     []string
		expected string
	}{
		{
			name: "skips element already in dst",
			dst: `
- name: foo
  image: foo:v2
`,
			src: `
- name: foo
  image: foo:v1
`,
			keys: []string{"name"},
			expected: `
- name: foo
  image: foo:v2
`,
		},
		{
			name: "appends element missing from dst",
			dst: `
- name: foo
  image: foo:v1
`,
			src: `
- name: bar
  image: bar:v1
`,
			keys: []string{"name"},
			expected: `
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
`,
		},
		{
			name: "mixed: skips existing, appends missing",
			dst: `
- name: foo
  image: foo:v2
`,
			src: `
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
`,
			keys: []string{"name"},
			expected: `
- name: foo
  image: foo:v2
- name: bar
  image: bar:v1
`,
		},
		{
			name: "scalar keys: skips existing values",
			dst: `
- a
- b
`,
			src: `
- b
- c
`,
			keys:     []string{""},
			expected: `
- a
- b
- c
`,
		},
		{
			name: "scalar keys: appends all missing",
			dst: `
- a
`,
			src: `
- b
- c
`,
			keys:     []string{""},
			expected: `
- a
- b
- c
`,
		},
		{
			name: "appends element with missing key",
			dst: `
- name: foo
  image: foo:v1
`,
			src: `
- image: bar:v1
`,
			keys: []string{"name"},
			expected: `
- name: foo
  image: foo:v1
- image: bar:v1
`,
		},
		{
			name: "multi-key: skips existing",
			dst: `
- containerPort: 8080
  protocol: TCP
`,
			src: `
- containerPort: 8080
  protocol: TCP
`,
			keys: []string{"containerPort", "protocol"},
			expected: `
- containerPort: 8080
  protocol: TCP
`,
		},
		{
			name: "multi-key: appends when one key differs",
			dst: `
- containerPort: 8080
  protocol: TCP
`,
			src: `
- containerPort: 8080
  protocol: UDP
`,
			keys: []string{"containerPort", "protocol"},
			expected: `
- containerPort: 8080
  protocol: TCP
- containerPort: 8080
  protocol: UDP
`,
		},
		{
			name: "empty src: no changes",
			dst: `
- name: foo
`,
			src:  `[]`,
			keys: []string{"name"},
			expected: `
- name: foo
`,
		},
		{
			name: "empty dst: appends all",
			dst:  `[]`,
			src: `
- name: foo
`,
			keys:     []string{"name"},
			expected: `[{name: foo}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := yaml.Parse(tt.dst)
			require.NoError(t, err)
			src, err := yaml.Parse(tt.src)
			require.NoError(t, err)
			expected, err := yaml.Parse(tt.expected)
			require.NoError(t, err)

			result, err := appendMissingListNode(dst, src, tt.keys)
			require.NoError(t, err)

			resultStr, err := result.String()
			require.NoError(t, err)
			expectedStr, err := expected.String()
			require.NoError(t, err)

			assert.Equal(t, expectedStr, resultStr)
		})
	}
}
