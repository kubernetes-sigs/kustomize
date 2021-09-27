// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveSeqIndentStyle(t *testing.T) {
	type testCase struct {
		name           string
		input          string
		expectedOutput string
	}

	testCases := []testCase{
		{
			name: "detect simple wide indent",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
`,
			expectedOutput: `wide`,
		},
		{
			name: "detect simple compact indent",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
`,
			expectedOutput: `compact`,
		},
		{
			name: "read with mixed indentation, wide first",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  - foo
  - bar
  - baz
env:
- foo
- bar
`,
			expectedOutput: `wide`,
		},
		{
			name: "read with mixed indentation, compact first",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
- baz
env:
  - foo
  - bar
`,
			expectedOutput: `compact`,
		},
		{
			name: "read with mixed indentation, compact first with less elements",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
- foo
- bar
env:
  - foo
  - bar
  - baz
`,
			expectedOutput: `compact`,
		},
		{
			name: "skip wrapped sequence strings, pipe hyphen",
			input: `apiVersion: apps/v1
kind: Deployment
spec: |-
  - foo
  - bar
`,
			expectedOutput: `compact`,
		},
		{
			name: "skip wrapped sequence strings, pipe",
			input: `apiVersion: apps/v1
kind: Deployment
spec: |
  - foo
  - bar
`,
			expectedOutput: `compact`,
		},
		{
			name: "skip wrapped sequence strings, right angle bracket",
			input: `apiVersion: apps/v1
kind: Deployment
spec: >
  - foo
  - bar
`,
			expectedOutput: `compact`,
		},
		{
			name: "skip wrapped sequence strings, plus",
			input: `apiVersion: apps/v1
kind: Deployment
spec: +
  - foo
  - bar
`,
			expectedOutput: `compact`,
		},
		{
			name: "handle comments",
			input: `apiVersion: v1
kind: Service
spec:
  ports: # comment 1
    # comment 2
    - name: etcd-server-ssl
      port: 2380
    # comment 3
    - name: etcd-client-ssl
      port: 2379
`,
			expectedOutput: `wide`,
		},
		{
			name: "nested wide vs compact",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  foo:
    bar:
      baz:
        bor:
          - a
          - b
abc:
- a
- b
`,
			expectedOutput: `wide`,
		},
		{
			name:           "invalid resource but valid yaml sequence",
			input:          `  - foo`,
			expectedOutput: `compact`,
		},
		{
			name: "invalid resource but valid yaml sequence with comments",
			input: `
# comment 1
  # comment 2
  - foo
`,
			expectedOutput: `compact`,
		},
		{
			name: "- within sequence element",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  foo:
    - - a`,
			expectedOutput: `wide`,
		},
		{
			name: "- within non sequence element",
			input: `apiVersion: apps/v1
kind: Deployment
spec:
  foo:
    a: - b`,
			expectedOutput: `compact`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedOutput, DeriveSeqIndentStyle(tc.input))
		})
	}
}
