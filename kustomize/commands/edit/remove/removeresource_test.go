// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/remove"
)

func TestRemoveResources(t *testing.T) {
	testCases := []remove.Case{
		{
			Description: "remove resources",
			Given: remove.Given{
				Items: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				RemoveArgs: []string{"resource1.yaml"},
			},
			Expected: remove.Expected{
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
			Given: remove.Given{
				Items: []string{
					"foo/resource1.yaml",
					"foo/resource2.yaml",
					"foo/resource3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{"foo/resource*.yaml"},
			},
			Expected: remove.Expected{
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
			Given: remove.Given{
				Items: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				RemoveArgs: []string{"foo"},
			},
			Expected: remove.Expected{
				Items: []string{
					"resource2.yaml",
					"resource3.yaml",
					"resource1.yaml",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       remove.Given{},
			Expected: remove.Expected{
				Err: errors.New("must specify a resource file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: remove.Given{
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
			Expected: remove.Expected{
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

	remove.ExecuteRemoveTestCases(t, testCases, "resources", newCmdRemoveResource)
}
