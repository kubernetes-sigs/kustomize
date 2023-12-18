// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/remove"
	"sigs.k8s.io/kustomize/kyaml/errors"
)

func TestRemoveTransformer(t *testing.T) {
	testCases := []remove.Case{
		{
			Description: "remove transformers",
			Given: remove.Given{
				Items: []string{
					"transformer1.yaml",
					"transformer2.yaml",
					"transformer3.yaml",
				},
				RemoveArgs: []string{"transformer1.yaml"},
			},
			Expected: remove.Expected{
				Items: []string{
					"transformer2.yaml",
					"transformer3.yaml",
				},
				Deleted: []string{
					"transformer1.yaml",
				},
			},
		},
		{
			Description: "remove transformer with pattern",
			Given: remove.Given{
				Items: []string{
					"foo/transformer1.yaml",
					"foo/transformer2.yaml",
					"foo/transformer3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{"foo/transformer*.yaml"},
			},
			Expected: remove.Expected{
				Items: []string{
					"do/not/deleteme/please.yaml",
				},
				Deleted: []string{
					"foo/transformer1.yaml",
					"foo/transformer2.yaml",
					"foo/transformer3.yaml",
				},
			},
		},
		{
			Description: "nothing found to remove",
			Given: remove.Given{
				Items: []string{
					"transformer1.yaml",
					"transformer2.yaml",
					"transformer3.yaml",
				},
				RemoveArgs: []string{"foo"},
			},
			Expected: remove.Expected{
				Items: []string{
					"transformer2.yaml",
					"transformer3.yaml",
					"transformer1.yaml",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       remove.Given{},
			Expected: remove.Expected{
				Err: errors.Errorf("must specify a transformer file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: remove.Given{
				Items: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"transformer3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{
					"foo/*.*",
					"bar/*.*",
					"tra*.yaml",
				},
			},
			Expected: remove.Expected{
				Items: []string{
					"do/not/deleteme/please.yaml",
				},
				Deleted: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"transformer3.yaml",
				},
			},
		},
	}

	remove.ExecuteRemoveTestCases(t, testCases, "transformers", newCmdRemoveTransformer)
}
