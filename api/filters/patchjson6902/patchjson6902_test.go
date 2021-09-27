// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patchjson6902

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
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
