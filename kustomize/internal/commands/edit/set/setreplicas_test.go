// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

func TestSetReplicas(t *testing.T) {
	type given struct {
		args           []string
		infileReplicas []string
	}
	type expected struct {
		fileOutput []string
		err        error
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			given: given{
				args: []string{"app=5"},
			},
			expected: expected{
				fileOutput: []string{
					"replicas:",
					"- count: 5",
					"  name: app",
				}},
		},
		{
			description: "override file",
			given: given{
				args: []string{"app=5"},
				infileReplicas: []string{
					"replicas:",
					"- count: 1",
					"  name: app",
					"- count: 2",
					"  name: other-app",
				},
			},
			expected: expected{
				fileOutput: []string{
					"replicas:",
					"- count: 5",
					"  name: app",
					"- count: 2",
					"  name: other-app",
				}},
		},
		{
			description: "multiple args with multiple overrides",
			given: given{
				args: []string{
					"app=100",
					"other-app=200",
				},
				infileReplicas: []string{
					"replicas:",
					"- count: 1",
					"  name: app",
					"- count: 2",
					"  name: other-app",
				},
			},
			expected: expected{
				fileOutput: []string{
					"replicas:",
					"- count: 100",
					"  name: app",
					"- count: 200",
					"  name: other-app",
				}},
		},
		{
			description: "error: no args",
			expected: expected{
				err: errReplicasNoArgs,
			},
		},
		{
			description: "error: invalid args -- no =",
			given: given{
				args: []string{"bad", "args"},
			},
			expected: expected{
				err: errReplicasInvalidArgs,
			},
		},
		{
			description: "error: invalid args -- non-integer count",
			given: given{
				args: []string{"app=bad"},
			},
			expected: expected{
				err: errReplicasInvalidArgs,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s%v", tc.description, tc.given.args), func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			cmd := newCmdSetReplicas(fSys)

			if len(tc.given.infileReplicas) > 0 {
				// write file with infileReplicas
				testutils_test.WriteTestKustomizationWith(
					fSys,
					[]byte(strings.Join(tc.given.infileReplicas, "\n")))
			} else {
				testutils_test.WriteTestKustomization(fSys)
			}

			// act
			err := cmd.RunE(cmd, tc.given.args)

			// assert
			if err != tc.expected.err {
				t.Errorf("Unexpected error from set replicas command. Actual: %v\nExpected: %v", err, tc.expected.err)
				t.FailNow()
			}

			content, err := testutils_test.ReadTestKustomization(fSys)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
				t.FailNow()
			}
			expectedStr := strings.Join(tc.expected.fileOutput, "\n")
			if !strings.Contains(string(content), expectedStr) {
				t.Errorf("unexpected replicas in kustomization file. \nActual:\n%s\nExpected:\n%s", content, expectedStr)
			}
		})
	}
}
