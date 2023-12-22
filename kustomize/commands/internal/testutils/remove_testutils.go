// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutils_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// RemoveTestGivenValues represents the provided inputs for the test case.
type RemoveTestGivenValues struct {
	// Items is the given input items.
	Items []string
	// RemoveArgs are the arguments to pass to the remove command.
	RemoveArgs []string
}

// RemoveTestExpectedValues represents the expected outputs of the test case.
type RemoveTestExpectedValues struct {
	// RemoveTestExpectedValues is the collection of expected output items.
	Items []string
	// Deleted is the collection of expected Deleted items (if any).
	Deleted []string
	// Err represents the error that is expected in the output (if any).
	Err error
}

// RemoveTestCase represents a test case to execute.
type RemoveTestCase struct {
	// Description is the description of the test case.
	Description string
	// Given is the provided inputs for the test case.
	Given RemoveTestGivenValues
	// Expected is the expected outputs for the test case.
	Expected RemoveTestExpectedValues
}

// ExecuteRemoveTestCases executes the provided test cases against the specified command
// for a particular collection (e.g. ) tests a command defined by the provided
// collection Name (e.g. transformers or resources) and newRemoveCmdToTest function.
func ExecuteRemoveTestCases(
	t *testing.T,
	testCases []RemoveTestCase,
	collectionName string,
	newRemoveCmdToTest func(filesys.FileSystem) *cobra.Command,
) {
	t.Helper()
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			WriteTestKustomizationWith(
				fSys,
				[]byte(fmt.Sprintf("%s:\n  - %s",
					collectionName,
					strings.Join(tc.Given.Items, "\n  - "))))
			cmd := newRemoveCmdToTest(fSys)
			err := cmd.RunE(cmd, tc.Given.RemoveArgs)
			if err != nil && tc.Expected.Err == nil {
				t.Errorf("unexpected cmd error: %v", err)
			} else if tc.Expected.Err != nil {
				if err.Error() != tc.Expected.Err.Error() {
					t.Errorf("expected error did not occur. Expected: %v. Actual: %v",
						tc.Expected.Err,
						err)
				}
				return
			}
			content, err := ReadTestKustomization(fSys)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
			}

			for _, itemFileName := range tc.Expected.Items {
				if !strings.Contains(string(content), itemFileName) {
					t.Errorf("expected item not found in kustomization file.\nItem: %s\nKustomization file:\n%s",
						itemFileName, content)
				}
			}
			for _, itemFileName := range tc.Expected.Deleted {
				if strings.Contains(string(content), itemFileName) {
					t.Errorf("expected deleted item found in kustomization file. Item: %s\nKustomization file:\n%s",
						itemFileName, content)
				}
			}
		})
	}
}
