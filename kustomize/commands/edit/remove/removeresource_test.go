// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v3/commands/edit/remove_test"
)

func TestRemoveResources(t *testing.T) {
	testCases := []remove_test.Case{
		{
			Description: "remove resources",
			Given: remove_test.Given{
				Items: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				RemoveArgs: []string{"resource1.yaml"},
			},
			Expected: remove_test.Expected{
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
			Given: remove_test.Given{
				Items: []string{
					"foo/resource1.yaml",
					"foo/resource2.yaml",
					"foo/resource3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{"foo/resource*.yaml"},
			},
			Expected: remove_test.Expected{
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
			Given: remove_test.Given{
				Items: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				RemoveArgs: []string{"foo"},
			},
			Expected: remove_test.Expected{
				Items: []string{
					"resource2.yaml",
					"resource3.yaml",
					"resource1.yaml",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       remove_test.Given{},
			Expected: remove_test.Expected{
				Err: errors.New("must specify a resource file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: remove_test.Given{
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
			Expected: remove_test.Expected{
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

	remove_test.ExecuteTestCases(t, testCases, "resources", newCmdRemoveResource)
}
