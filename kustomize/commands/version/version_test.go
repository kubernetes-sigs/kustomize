// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunInvalidOutput(t *testing.T) {
	var buf bytes.Buffer
	o := NewOptions(&buf)
	o.Output = "yml"
	err := o.Run()
	if err == nil {
		t.Fatal("expected error for invalid --output")
	}
	if !strings.Contains(err.Error(), "--output must be 'yaml' or 'json'") {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output on invalid --output, got %q", buf.String())
	}
}
