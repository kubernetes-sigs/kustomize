// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package imagetag

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestLegacyImageTag_Filter(t *testing.T) {
	testCases := map[string]struct {
		input          string
		expectedOutput string
		filter         LegacyFilter
	}{
		"updates multiple images inside containers": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: nginx:2.1.2
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache@12345
  - image: apache@12345
`,
			filter: LegacyFilter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					Digest:  "12345",
				},
			},
		},
		"updates inside both containers and initContainers": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: tomcat:1.2.3
  initContainers:
  - image: nginx:1.2.1
  - image: apache:1.2.3
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache:3.2.1
  - image: tomcat:1.2.3
  initContainers:
  - image: apache:3.2.1
  - image: apache:1.2.3
`,
			filter: LegacyFilter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					NewTag:  "3.2.1",
				},
			},
		},
		"updates on multiple depths": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: tomcat:1.2.3
  template:
    spec: 
      initContainers:
      - image: nginx:1.2.1
      - image: apache:1.2.3
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache:3.2.1
  - image: tomcat:1.2.3
  template:
    spec:
      initContainers:
      - image: apache:3.2.1
      - image: apache:1.2.3
`,
			filter: LegacyFilter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					NewTag:  "3.2.1",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			filter := tc.filter
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(filtertest.RunFilter(t, tc.input, filter))) {
				t.FailNow()
			}
		})
	}
}
