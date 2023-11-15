package set

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestNewSetGeneratorOptionsIsNotNil(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	assert.NotNil(t, newCmdSetGeneratorOptions(fSys), nil)
}

func TestRunSetGeneratorOptionsValidate(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	cmd := newCmdSetGeneratorOptions(fSys)
	testutils_test.WriteTestKustomization(fSys)
	args := []string{}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, "at least immutable, or disableNameSuffixHash or labels or annotations must be set.", err.Error())
}

func TestRunSetGeneratorOptionsWithImmutable(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	cmd := newCmdSetGeneratorOptions(fSys)
	testutils_test.WriteTestKustomization(fSys)
	args := []string{"--immutable"}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	expectedStr := strings.Join([]string{
		"generatorOptions:",
		"  immutable: true",
	}, "\n")
	assert.Contains(t, string(content), expectedStr)
}

func TestRunSetGeneratorOptionsWithDisableNameSuffixHash(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	cmd := newCmdSetGeneratorOptions(fSys)
	testutils_test.WriteTestKustomization(fSys)
	args := []string{"--disableNameSuffixHash"}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	expectedStr := strings.Join([]string{
		"generatorOptions:",
		"  disableNameSuffixHash: true",
	}, "\n")
	assert.Contains(t, string(content), expectedStr)
}

func TestRunSetGeneratorOptionsWithLabels(t *testing.T) {
	type given struct {
		args []string
	}
	type expected struct {
		fileOutput []string
		errMessage string
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			description: "One pair",
			given: given{
				args: []string{
					"--labels=a:b",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  labels:",
					"    a: b",
				},
			},
		},
		{
			description: "Two pairs",
			given: given{
				args: []string{
					"--labels=a:b",
					"--labels=c:d",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  labels:",
					"    a: b",
					"    c: d",
				},
			},
		},
		{
			description: "There are three colons",
			given: given{
				args: []string{
					"--labels=a:b:c",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  labels:",
					"    a: b:c",
				},
			},
		},
		{
			description: "Empty value",
			given: given{
				args: []string{
					"--labels=a: ",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  labels:",
					"    a:",
				},
			},
		},
		{
			description: "No colon",
			given: given{
				args: []string{
					"--labels=x,y",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  labels:",
					"    x,y:",
				},
			},
		},
		{
			description: "Empty key",
			given: given{
				args: []string{
					"--labels=:y",
				},
			},
			expected: expected{
				errMessage: "invalid label: ':y' (need k:v pair where v may be quoted)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s%v", tc.description, tc.given.args), func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			cmd := newCmdSetGeneratorOptions(fSys)
			testutils_test.WriteTestKustomization(fSys)
			cmd.SetArgs(tc.given.args)

			err := cmd.Execute()

			errMessage := ""
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.expected.errMessage {
				t.Errorf("Unexpected error from set generatoroptions command. Actual: %s\nExpected: %s", errMessage, tc.expected.errMessage)
				t.FailNow()
			}

			content, err := testutils_test.ReadTestKustomization(fSys)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
				t.FailNow()
			}

			expectedStr := strings.Join(tc.expected.fileOutput, "\n")
			if !strings.Contains(string(content), expectedStr) {
				t.Errorf("unexpected image in kustomization file. \nActual:\n%s\nExpected:\n%s", content, expectedStr)
			}
		})
	}
}

func TestRunSetGeneratorOptionsWithAnnotations(t *testing.T) {
	type given struct {
		args []string
	}
	type expected struct {
		fileOutput []string
		errMessage string
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			description: "One pair",
			given: given{
				args: []string{
					"--annotations=a:b",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  annotations:",
					"    a: b",
				},
			},
		},
		{
			description: "Two pairs",
			given: given{
				args: []string{
					"--annotations=a:b",
					"--annotations=c:d",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  annotations:",
					"    a: b",
					"    c: d",
				},
			},
		},
		{
			description: "There are three colons",
			given: given{
				args: []string{
					"--annotations=a:b:c",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  annotations:",
					"    a: b:c",
				},
			},
		},
		{
			description: "Empty value",
			given: given{
				args: []string{
					"--annotations=a: ",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  annotations:",
					"    a:",
				},
			},
		},
		{
			description: "No colon",
			given: given{
				args: []string{
					"--annotations=x,y",
				},
			},
			expected: expected{
				fileOutput: []string{
					"generatorOptions:",
					"  annotations:",
					"    x,y:",
				},
			},
		},
		{
			description: "Empty key",
			given: given{
				args: []string{
					"--annotations=:y",
				},
			},
			expected: expected{
				errMessage: "invalid annotation: ':y' (need k:v pair where v may be quoted)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s%v", tc.description, tc.given.args), func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			cmd := newCmdSetGeneratorOptions(fSys)
			testutils_test.WriteTestKustomization(fSys)
			cmd.SetArgs(tc.given.args)

			err := cmd.Execute()

			errMessage := ""
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.expected.errMessage {
				t.Errorf("Unexpected error from set generatoroptions command. Actual: %s\nExpected: %s", errMessage, tc.expected.errMessage)
				t.FailNow()
			}

			content, err := testutils_test.ReadTestKustomization(fSys)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
				t.FailNow()
			}

			expectedStr := strings.Join(tc.expected.fileOutput, "\n")
			if !strings.Contains(string(content), expectedStr) {
				t.Errorf("unexpected image in kustomization file. \nActual:\n%s\nExpected:\n%s", content, expectedStr)
			}
		})
	}
}
