package remove_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

// Given represents the provided inputs for the test case.
type Given struct {
	// Items is the given input items.
	Items []string
	// RemoveArgs are the arguments to pass to the remove command.
	RemoveArgs []string
}

// Expected represents the expected outputs of the test case.
type Expected struct {
	// Expected is the collection of expected output items.
	Items []string
	// Deleted is the collection of expected Deleted items (if any).
	Deleted []string
	// Err represents the error that is expected in the output (if any).
	Err error
}

// Case represents a test case to execute.
type Case struct {
	// Description is the description of the test case.
	Description string
	// Given is the provided inputs for the test case.
	Given Given
	// Expected is the expected outputs for the test case.
	Expected Expected
}

// ExecuteTestCases executes the provided test cases against the specified command
// for a particular collection (e.g. ) tests a command defined by the provided
// collection Name (e.g. transformers or resources) and newRemoveCmdToTest function.
func ExecuteTestCases(t *testing.T, testCases []Case, collectionName string,
	newRemoveCmdToTest func(filesys.FileSystem) *cobra.Command) {
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(
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
					t.Errorf("expected error did not occurred. Expected: %v. Actual: %v",
						tc.Expected.Err,
						err)
				}
				return
			}
			content, err := testutils_test.ReadTestKustomization(fSys)
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
