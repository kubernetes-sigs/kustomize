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

func TestImageTagUpdater_Filter(t *testing.T) {
	testCases := map[string]struct {
		input          string
		expectedOutput string
		filter         Filter
		fsSlice        types.FsSlice
	}{
		"update with digest": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  image: nginx:1.2.1
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  image: apache@12345
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					Digest:  "12345",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/image",
				},
			},
		},
		"multiple matches in sequence": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: not_nginx@54321
  - image: nginx:1.2.1
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache:3.2.1
  - image: not_nginx@54321
  - image: apache:3.2.1
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					NewTag:  "3.2.1",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers/image",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			filter := tc.filter
			filter.FsSlice = tc.fsSlice
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(filtertest.RunFilter(t, tc.input, filter))) {
				t.FailNow()
			}
		})
	}
}
