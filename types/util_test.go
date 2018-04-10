/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestFilterByGVK(t *testing.T) {
	type testCase struct {
		description string
		in          schema.GroupVersionKind
		filter      *schema.GroupVersionKind
		expected    bool
	}
	testCases := []testCase{
		{
			description: "nil filter",
			in:          schema.GroupVersionKind{},
			filter:      nil,
			expected:    true,
		},
		{
			description: "GVK matches",
			in: schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			expected: true,
		},
		{
			description: "group doesn't matches",
			in: schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &schema.GroupVersionKind{
				Group:   "group2",
				Version: "version1",
				Kind:    "kind1",
			},
			expected: false,
		},
		{
			description: "version doesn't matches",
			in: schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &schema.GroupVersionKind{
				Group:   "group1",
				Version: "version2",
				Kind:    "kind1",
			},
			expected: false,
		},
		{
			description: "kind doesn't matches",
			in: schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind2",
			},
			expected: false,
		},
		{
			description: "no version in filter",
			in: schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &schema.GroupVersionKind{
				Group:   "group1",
				Version: "",
				Kind:    "kind1",
			},
			expected: true,
		},
		{
			description: "only kind is set in filter",
			in: schema.GroupVersionKind{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &schema.GroupVersionKind{
				Group:   "",
				Version: "",
				Kind:    "kind1",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		filtered := SelectByGVK(tc.in, tc.filter)
		if filtered != tc.expected {
			t.Fatalf("unexpected filter result for test case: %v", tc.description)
		}
	}
}
