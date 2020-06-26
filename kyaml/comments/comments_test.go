// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package comments

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCopyComments(t *testing.T) {
	testCases := []struct {
		name     string
		from     string
		to       string
		expected string
	}{
		{
			name: "copy_comments",
			from: `# A
#
# B

# C
apiVersion: apps/v1
kind: Deployment
spec: # comment 1
  # comment 2
  replicas: 3 # comment 3
  # comment 4
`,
			to: `apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 4
`,
			expected: `# A
#
# B

# C
apiVersion: apps/v1
kind: Deployment
spec: # comment 1
  # comment 2
  replicas: 4 # comment 3
  # comment 4
`,
		},

		{
			name: "associative_list",
			from: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: bar # comment 1
`,
			to: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: bar
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: bar # comment 1
`,
		},

		{
			name: "keep_comments",
			from: `# A
#
# B

# C
apiVersion: apps/v1
kind: Deployment
spec: # comment 1
  # comment 2
  replicas: 3 # comment 3
  # comment 4
`,
			to: `apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 4 # comment 5
`,
			expected: `# A
#
# B

# C
apiVersion: apps/v1
kind: Deployment
spec: # comment 1
  # comment 2
  replicas: 4 # comment 5
  # comment 4
`,
		},

		{
			name: "copy_item_comments",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- a
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
`,
		},

		{
			name: "copy_item_comments_2",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
# comment
- a
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- a
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
# comment
- a
`,
		},

		{
			name: "copy_item_comments_middle",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
- a
- b # comment
- c
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- d
- b
- e
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
- d
- b # comment
- e
`,
		},

		{
			name: "copy_item_comments_moved",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
- a
- b # comment
- c
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- a
- c
- b
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
- a
- c
- b
`,
		},

		{
			name: "copy_item_comments_no_match",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- b
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
- b
`,
		},

		{
			name: "copy_item_comments_add",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- a
- b
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
- b
`,
		},

		{
			name: "copy_item_comments_remove",
			from: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
- b
`,
			to: `
apiVersion: apps/v1
kind: Deployment
items:
- a
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
items:
- a # comment
`,
		},

		{
			name: "copy_comments_folded_style",
			from: `
apiVersion: v1
kind: ConfigMap
data:
  somekey: "012345678901234567890123456789012345678901234567890123456789012345678901234" # x
`,
			to: `
apiVersion: v1
kind: ConfigMap
data:
  somekey: >-
    012345678901234567890123456789012345678901234567890123456789012345678901234
`,
			expected: `
apiVersion: v1
kind: ConfigMap
data:
  somekey: "012345678901234567890123456789012345678901234567890123456789012345678901234" # x
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			from, err := yaml.Parse(tc.from)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			to, err := yaml.Parse(tc.to)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = CopyComments(from, to)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := to.String()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}
