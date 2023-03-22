// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v5/commands/edit/remove_test"
)

func TestRemoveComponents(t *testing.T) {
	testCases := []remove_test.Case{
		{
			Description: "remove components",
			Given: remove_test.Given{
				Items: []string{
					"../../components/component1",
					"../../components/component2",
					"../../components/component3",
				},
				RemoveArgs: []string{"../../components/component2"},
			},
			Expected: remove_test.Expected{
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
			Given: remove_test.Given{
				Items: []string{
					"../../component/component1",
					"../../component/component2",
					"../../component/component3",
					"../../component/do_not_delete",
				},
				RemoveArgs: []string{"../../component/component*"},
			},
			Expected: remove_test.Expected{
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
			Given: remove_test.Given{
				Items: []string{
					"component/component1",
					"component/component2",
					"component/component3",
				},
				RemoveArgs: []string{"component/component4"},
			},
			Expected: remove_test.Expected{
				Items: []string{
					"component/component1",
					"component/component2",
					"component/component3",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       remove_test.Given{},
			Expected: remove_test.Expected{
				Err: errors.New("must specify a component"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: remove_test.Given{
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
			Expected: remove_test.Expected{
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

	remove_test.ExecuteTestCases(t, testCases, "components", newCmdRemoveComponent)
}
