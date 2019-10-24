// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

func TestRemoveResources(t *testing.T) {
	type given struct {
		resources  []string
		removeArgs []string
	}
	type expected struct {
		resources []string
		deleted   []string
		err       error
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			description: "remove resource",
			given: given{
				resources: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				removeArgs: []string{"resource1.yaml"},
			},
			expected: expected{
				resources: []string{
					"resource2.yaml",
					"resource3.yaml",
				},
				deleted: []string{
					"resource1.yaml",
				},
			},
		},
		{
			description: "remove resources with pattern",
			given: given{
				resources: []string{
					"foo/resource1.yaml",
					"foo/resource2.yaml",
					"foo/resource3.yaml",
					"do/not/deleteme/please.yaml",
				},
				removeArgs: []string{"foo/resource*.yaml"},
			},
			expected: expected{
				resources: []string{
					"do/not/deleteme/please.yaml",
				},
				deleted: []string{
					"foo/resource1.yaml",
					"foo/resource2.yaml",
					"foo/resource3.yaml",
				},
			},
		},
		{
			description: "nothing found to remove",
			given: given{
				resources: []string{
					"resource1.yaml",
					"resource2.yaml",
					"resource3.yaml",
				},
				removeArgs: []string{"foo"},
			},
			expected: expected{
				resources: []string{
					"resource2.yaml",
					"resource3.yaml",
					"resource1.yaml",
				},
			},
		},
		{
			description: "no arguments",
			given:       given{},
			expected: expected{
				err: errors.New("must specify a resource file"),
			},
		},
		{
			description: "remove with multiple pattern arguments",
			given: given{
				resources: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"resource3.yaml",
					"do/not/deleteme/please.yaml",
				},
				removeArgs: []string{
					"foo/*.*",
					"bar/*.*",
					"res*.yaml",
				},
			},
			expected: expected{
				resources: []string{
					"do/not/deleteme/please.yaml",
				},
				deleted: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"resource3.yaml",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(
				fSys,
				[]byte(fmt.Sprintf(
					"resources:\n  - %s", strings.Join(tc.given.resources, "\n  - "))))
			cmd := newCmdRemoveResource(fSys)
			err := cmd.RunE(cmd, tc.given.removeArgs)
			if err != nil && tc.expected.err == nil {

				t.Errorf("unexpected cmd error: %v", err)
			} else if tc.expected.err != nil {
				if err.Error() != tc.expected.err.Error() {
					t.Errorf("expected error did not occurred. Expected: %v. Actual: %v", tc.expected.err, err)
				}
				return
			}
			content, err := testutils_test.ReadTestKustomization(fSys)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
			}

			for _, resourceFileName := range tc.expected.resources {
				if !strings.Contains(string(content), resourceFileName) {
					t.Errorf("expected resource not found in kustomization file.\nResource: %s\nKustomization file:\n%s", resourceFileName, content)
				}
			}
			for _, resourceFileName := range tc.expected.deleted {
				if strings.Contains(string(content), resourceFileName) {
					t.Errorf("expected deleted resource found in kustomization file. Resource: %s\nKustomization file:\n%s", resourceFileName, content)
				}
			}

		})
	}
}
