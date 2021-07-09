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
			name: "skip wrapped sequence strings",
			input: `apiVersion: apps/v1
kind: Deployment
spec: |-
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
