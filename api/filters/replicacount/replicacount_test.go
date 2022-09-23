// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replicacount

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter(t *testing.T) {
	mutationTrackerStub := filtertest_test.MutationTrackerStub{}
	testCases := map[string]struct {
		input                string
		expected             string
		filter               Filter
		mutationTracker      func(key, value, tag string, node *yaml.RNode)
		expectedSetValueArgs []filtertest_test.SetValueArg
	}{
		"update field": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 5
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "dep",
					Count: 42,
				},
				FieldSpec: types.FieldSpec{Path: "spec/replicas"},
			},
		},
		"add field": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
    replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
				FieldSpec: types.FieldSpec{
					Path:               "spec/template/replicas",
					CreateIfNotPresent: true,
				},
			},
		},

		"add_field_null": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
    replicas: null
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
    replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
				FieldSpec: types.FieldSpec{
					Path:               "spec/template/replicas",
					CreateIfNotPresent: true,
				},
			},
		},
		"no update if CreateIfNotPresent is false": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
				FieldSpec: types.FieldSpec{
					Path: "spec/template/replicas",
				},
			},
		},
		"update multiple fields": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    replicas: 5
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
				FieldSpec: types.FieldSpec{Path: "spec/template/replicas"},
			},
		},
		"mutation tracker": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 5
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "dep",
					Count: 42,
				},
				FieldSpec: types.FieldSpec{Path: "spec/replicas"},
			},
			mutationTracker: mutationTrackerStub.MutationTracker,
			expectedSetValueArgs: []filtertest_test.SetValueArg{
				{
					Value:    "42",
					Tag:      "!!int",
					NodePath: []string{"spec", "replicas"},
				},
			},
		},
	}

	for tn, tc := range testCases {
		mutationTrackerStub.Reset()
		tc.filter.WithMutationTracker(tc.mutationTracker)
		t.Run(tn, func(t *testing.T) {
			if !assert.Equal(t,
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, tc.input, tc.filter))) {
				t.FailNow()
			}
			assert.Equal(t, tc.expectedSetValueArgs, mutationTrackerStub.SetValueArgs())
		})
	}
}
