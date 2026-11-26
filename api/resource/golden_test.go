// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use flag.Lookup to check if update-golden flag is set
// This avoids flag name conflicts and global variable linting issues
func isUpdateGolden() bool {
	f := flag.Lookup("update-golden")
	return f != nil && f.Value.String() == "true"
}

// goldenFile returns the path to the golden file for the given test name.
func goldenFile(t *testing.T, name string) string {
	t.Helper()
	name = strings.ReplaceAll(name, " ", "_")
	return filepath.Join("testdata", "golden", name+".golden")
}

// assertGoldenYAML compares the actual YAML output with the golden file.
// It automatically uses the test name from t.Name().
// If -update-golden flag is set, it updates the golden file instead.
func assertGoldenYAML(t *testing.T, actual []byte) {
	t.Helper()
	assertGolden(t, t.Name(), actual)
}

// assertGolden compares the actual output with the golden file.
// If -update-golden flag is set, it updates the golden file instead.
func assertGolden(t *testing.T, name string, actual []byte) {
	t.Helper()
	goldenPath := goldenFile(t, name)

	if isUpdateGolden() {
		// Update golden file
		dir := filepath.Dir(goldenPath)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(goldenPath, actual, 0644))
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
	assert.Equal(t, string(expected), string(actual), "Output does not match golden file. Run with -update-golden to update it.")
}
