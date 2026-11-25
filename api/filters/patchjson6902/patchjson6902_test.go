// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patchjson6902

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const input = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`

func TestPreservesLargeIntegers(t *testing.T) {
	node := yaml.MustParse(`
apiVersion: example.com/v1
kind: Example
metadata:
  name: example
spec:
  id: 9007199254740993
  region: original
`)
	_, err := (Filter{Patch: `[
{"op":"replace","path":"/spec/region","value":"west"}
]`}).Filter([]*yaml.RNode{node})
	require.NoError(t, err)
	id, err := node.Pipe(yaml.Lookup("spec", "id"))
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, "9007199254740993", id.YNode().Value)
	assert.Equal(t, yaml.NodeTagInt, id.YNode().Tag)
	region, err := node.Pipe(yaml.Lookup("spec", "region"))
	require.NoError(t, err)
	require.NotNil(t, region)
	assert.Equal(t, "west", region.YNode().Value)

	// A later numeric test must see the exact value left by the first patch.
	_, err = (Filter{Patch: `[
{"op":"test","path":"/spec/id","value":9007199254740993}
]`}).Filter([]*yaml.RNode{node})
	require.NoError(t, err)
	_, err = (Filter{Patch: `[
{"op":"test","path":"/spec/id","value":9007199254740992}
]`}).Filter([]*yaml.RNode{node})
	require.Error(t, err)
}

func TestSomething(t *testing.T) {
	testCases := []struct {
		testName       string
		input          string
		filter         Filter
		expectedOutput string
	}{
		{
			testName: "single operation, json",
			input:    input,
			filter: Filter{
				Patch: `[
{"op": "replace", "path": "/spec/replica", "value": 5}
]`,
			},
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 5
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`,
		},
		{
			testName: "multiple operations, json",
			input:    input,
			filter: Filter{
				Patch: `[
{"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
{"op": "add", "path": "/spec/replica", "value": 999},
{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`,
			},
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 999
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - command:
        - arg1
        - arg2
        - arg3
        image: nginx
        name: my-nginx
`,
		},
		{
			testName: "single operation, yaml",
			input:    input,
			filter: Filter{
				Patch: `
- op: replace
  path: /spec/replica
  value: 5
`,
			},
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 5
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`,
		},
		{
			testName: "multiple operations, yaml",
			input:    input,
			filter: Filter{
				Patch: `
- op: replace
  path: /spec/template/spec/containers/0/name
  value: my-nginx
- op: add
  path: /spec/replica
  value: 999
- op: add
  path: /spec/template/spec/containers/0/command
  value:
  - arg1
  - arg2
  - arg3
`,
			},
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 999
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - command:
        - arg1
        - arg2
        - arg3
        image: nginx
        name: my-nginx
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(
					filtertest.RunFilter(t, tc.input, tc.filter))) {
				t.FailNow()
			}
		})
	}
}
