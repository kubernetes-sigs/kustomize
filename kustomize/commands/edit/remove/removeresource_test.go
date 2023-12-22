// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"testing"

	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
)

func TestRemoveResources(t *testing.T) {
	testCases := []testutils_test.RemoveTestCase{
		{
			Description: "remove resources",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				RemoveArgs: []string{"resource1.yaml"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"resource2.yaml",
					"resource3.yaml",
				},
				Deleted: []string{
					"resource1.yaml",
				},
			},
		},
		{
			Description: "remove resource with pattern",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"foo/resource1.yaml",
					"foo/resource2.yaml",
					"foo/resource3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{"foo/resource*.yaml"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"do/not/deleteme/please.yaml",
				},
				Deleted: []string{
					"foo/resource1.yaml",
					"foo/resource2.yaml",
					"foo/resource3.yaml",
				},
			},
		},
		{
			Description: "nothing found to remove",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				RemoveArgs: []string{"foo"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"resource2.yaml",
					"resource3.yaml",
					"resource1.yaml",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       testutils_test.RemoveTestGivenValues{},
			Expected: testutils_test.RemoveTestExpectedValues{
				Err: errors.New("must specify a resource file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"resource3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{
					"foo/*.*",
					"bar/*.*",
					"res*.yaml",
				},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"do/not/deleteme/please.yaml",
				},
				Deleted: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"resource3.yaml",
				},
			},
		},
	}

	testutils_test.ExecuteRemoveTestCases(t, testCases, "resources", newCmdRemoveResource)
}
