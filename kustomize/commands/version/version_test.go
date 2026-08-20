// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"bytes"
	"testing"
)

func TestVersionInvalidOutput(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCmdVersion(&buf)
	cmd.SetArgs([]string{"--output", "yml"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid output format, got nil")
	}
	expectedErr := "--output must be 'yaml' or 'json'"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestVersionValidOutput(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCmdVersion(&buf)
	cmd.SetArgs([]string{"--output", "yaml"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error for valid output format: %v", err)
	}
}
