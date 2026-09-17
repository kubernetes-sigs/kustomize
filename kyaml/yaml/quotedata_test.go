// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestQuoteStringDataMapsInYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "quotes every configmap data value",
			// Matches kubernetes-sigs/kustomize#5558: ConfigMap data is
			// map[string]string, so every value must remain a quoted string.
			input: `apiVersion: v1
data:
  test1: ${TEMPLATE_VAR1}
  test2: '{TEMPLATE_VAR2}'
  test3: "true"
  test4: test4
kind: ConfigMap
metadata:
  name: config
`,
			expected: `apiVersion: v1
data:
  test1: "${TEMPLATE_VAR1}"
  test2: '{TEMPLATE_VAR2}'
  test3: "true"
  test4: "test4"
kind: ConfigMap
metadata:
  name: config
`,
		},
		{
			name: "quotes dollar placeholders from the original report",
			input: `apiVersion: v1
data:
  pci: ${TEST}
kind: ConfigMap
metadata:
  name: test-object
`,
			expected: `apiVersion: v1
data:
  pci: "${TEST}"
kind: ConfigMap
metadata:
  name: test-object
`,
		},
		{
			name: "keeps multiline configmap values as block scalars",
			input: `apiVersion: v1
data:
  nginx.conf: |
    server {
      listen 80;
    }
kind: ConfigMap
metadata:
  name: nginx
`,
			expected: `apiVersion: v1
data:
  nginx.conf: |
    server {
      listen 80;
    }
kind: ConfigMap
metadata:
  name: nginx
`,
		},
		{
			name: "quotes secret stringData values",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: s
stringData:
  token: ${TOKEN}
  user: admin
`,
			expected: `apiVersion: v1
kind: Secret
metadata:
  name: s
stringData:
  token: "${TOKEN}"
  user: "admin"
`,
		},
		{
			name: "leaves deployments unchanged",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
spec:
  replicas: 1
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
spec:
  replicas: 1
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := yaml.QuoteStringDataMapsInYAML([]byte(tc.input))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(got))
		})
	}
}

func TestHasStringDataMaps(t *testing.T) {
	cm, err := yaml.Parse("kind: ConfigMap\n")
	require.NoError(t, err)
	assert.True(t, yaml.HasStringDataMaps(cm))

	dep, err := yaml.Parse("kind: Deployment\n")
	require.NoError(t, err)
	assert.False(t, yaml.HasStringDataMaps(dep))
}
