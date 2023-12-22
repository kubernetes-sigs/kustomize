// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"testing"

	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/errors"
)

func TestRemoveTransformer(t *testing.T) {
	testCases := []testutils_test.RemoveTestCase{
		{
			Description: "remove transformers",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"transformer1.yaml",
					"transformer2.yaml",
					"transformer3.yaml",
				},
				RemoveArgs: []string{"transformer1.yaml"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
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
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"foo/transformer1.yaml",
					"foo/transformer2.yaml",
					"foo/transformer3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{"foo/transformer*.yaml"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
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
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"transformer1.yaml",
					"transformer2.yaml",
					"transformer3.yaml",
				},
				RemoveArgs: []string{"foo"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"transformer2.yaml",
					"transformer3.yaml",
					"transformer1.yaml",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       testutils_test.RemoveTestGivenValues{},
			Expected: testutils_test.RemoveTestExpectedValues{
				Err: errors.Errorf("must specify a transformer file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: testutils_test.RemoveTestGivenValues{
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
			Expected: testutils_test.RemoveTestExpectedValues{
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

	testutils_test.ExecuteRemoveTestCases(t, testCases, "transformers", newCmdRemoveTransformer)
}
