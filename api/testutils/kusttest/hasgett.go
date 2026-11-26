// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/resmap"
)

// updateGoldenFlagName is the flag name for updating golden files.
// We use flag.Lookup() instead of a global variable to avoid linting issues.
const updateGoldenFlagName = "update-golden"

// isUpdateGolden checks if the update-golden flag is set.
// The flag is registered and parsed in TestMain (see testmain_test.go).
func isUpdateGolden() bool {
	f := flag.Lookup(updateGoldenFlagName)
	return f != nil && f.Value.String() == "true"
}

type hasGetT interface {
	GetT() *testing.T
}

func assertActualEqualsExpectedWithTweak(
	ht hasGetT,
	m resmap.ResMap,
	tweaker func([]byte) []byte, expected string) {
	AssertActualEqualsExpectedWithTweak(ht.GetT(), m, tweaker, expected)
}

func AssertActualEqualsExpectedWithTweak(
	t *testing.T,
	m resmap.ResMap,
	tweaker func([]byte) []byte, expected string) {
	t.Helper()
	if m == nil {
		t.Fatalf("Map should not be nil.")
	}
	// Ignore leading linefeed in expected value
	// to ease readability of tests.
	if len(expected) > 0 && expected[0] == 10 {
		expected = expected[1:]
	}
	actual, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	if tweaker != nil {
		actual = tweaker(actual)
	}

	// Use golden file if update flag is set or golden file exists
	goldenPath := goldenFileForTest(t)
	if isUpdateGolden() {
		// Update golden file
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create golden file directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Try to read golden file first
	goldenBytes, err := os.ReadFile(goldenPath)
	if err == nil {
		// Golden file exists, use it
		if string(goldenBytes) != string(actual) {
			reportDiffAndFail(t, actual, string(goldenBytes))
		}
		return
	}

	// Only fall back to expected string if golden file doesn't exist
	// (backward compatibility). Fail loudly on other I/O errors.
	if !os.IsNotExist(err) {
		t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
	}

	// Fall back to expected string (backward compatibility)
	if string(actual) != expected {
		reportDiffAndFail(t, actual, expected)
	}
}

// goldenFileForTest returns the path to the golden file for the current test.
func goldenFileForTest(t *testing.T) string {
	t.Helper()
	// Use test name as golden file name
	// Replace / with _ for subtest names
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	return filepath.Join("testdata", "golden", testName+".golden")
}

// AssertYAMLEqualsGolden compares the actual YAML bytes with a golden file.
// If -update-golden flag is set, it updates the golden file instead.
// This is for tests that use assert.Equal directly with AsYaml() output.
func AssertYAMLEqualsGolden(t *testing.T, actual []byte) {
	t.Helper()
	goldenPath := goldenFileForTest(t)

	if isUpdateGolden() {
		// Update golden file
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create golden file directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read golden file
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file does not exist: %s\nRun tests with -update-golden flag to create it.", goldenPath)
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare
	if string(expected) != string(actual) {
		reportDiffAndFail(t, actual, string(expected))
	}
}

// Pretty printing of file differences.
func reportDiffAndFail(
	t *testing.T, actual []byte, expected string) {
	t.Helper()
	sE, maxLen := convertToArray(expected)
	sA, _ := convertToArray(string(actual))
	fmt.Println("===== ACTUAL BEGIN ========================================")
	fmt.Print(string(actual))
	fmt.Println("===== ACTUAL END ==========================================")
	format := fmt.Sprintf("%%s  %%-%ds %%s\n", maxLen+4)
	var limit int
	if len(sE) < len(sA) {
		limit = len(sE)
	} else {
		limit = len(sA)
	}
	fmt.Printf(format, " ", "EXPECTED", "ACTUAL")
	fmt.Printf(format, " ", "--------", "------")
	for i := 0; i < limit; i++ {
		fmt.Printf(format, hint(sE[i], sA[i]), sE[i], sA[i])
	}
	if len(sE) < len(sA) {
		for i := len(sE); i < len(sA); i++ {
			fmt.Printf(format, "X", "", sA[i])
		}
	} else {
		for i := len(sA); i < len(sE); i++ {
			fmt.Printf(format, "X", sE[i], "")
		}
	}
	t.Fatalf("Expected not equal to actual")
}

func hint(a, b string) string {
	if a == b {
		return " "
	}
	return "X"
}

func convertToArray(x string) ([]string, int) {
	a := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	maxLen := 0
	for i, v := range a {
		z := tabToSpace(v)
		if len(z) > maxLen {
			maxLen = len(z)
		}
		a[i] = z
	}
	return a, maxLen
}

func tabToSpace(input string) string {
	var result []string
	for _, i := range input {
		if i == 9 {
			result = append(result, "  ")
		} else {
			result = append(result, string(i))
		}
	}
	return strings.Join(result, "")
}
