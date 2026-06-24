// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"testing"

	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
)

func TestRemoveComponents(t *testing.T) {
	testCases := []testutils_test.RemoveTestCase{
		{
			Description: "remove components",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"../../components/component1",
					"../../components/component2",
					"../../components/component3",
				},
				RemoveArgs: []string{"../../components/component2"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"../../components/component1",
					"../../components/component3",
				},
				Deleted: []string{
					"../../components/component2",
				},
			},
		},
		{
			Description: "remove component with pattern",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"../../component/component1",
					"../../component/component2",
					"../../component/component3",
					"../../component/do_not_delete",
				},
				RemoveArgs: []string{"../../component/component*"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"../../component/do_not_delete",
				},
				Deleted: []string{
					"../../component/component1",
					"../../component/component2",
					"../../component/component3",
				},
			},
		},
		{
			Description: "nothing found to remove",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"component/component1",
					"component/component2",
					"component/component3",
				},
				RemoveArgs: []string{"component/component4"},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"component/component1",
					"component/component2",
					"component/component3",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       testutils_test.RemoveTestGivenValues{},
			Expected: testutils_test.RemoveTestExpectedValues{
				Err: errors.New("must specify a component file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: testutils_test.RemoveTestGivenValues{
				Items: []string{
					"foo/component1",
					"bar/component2",
					"do_not_delete",
					"component3",
				},
				RemoveArgs: []string{
					"foo/*",
					"bar/*",
					"compo*",
				},
			},
			Expected: testutils_test.RemoveTestExpectedValues{
				Items: []string{
					"do_not_delete",
				},
				Deleted: []string{
					"foo/component1",
					"bar/component2",
					"component3",
				},
			},
		},
	}

	testutils_test.ExecuteRemoveTestCases(t, testCases, "components", newCmdRemoveComponent)
}
