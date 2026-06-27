// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version_test

import (
	"bytes"
	"testing"

	. "sigs.k8s.io/kustomize/kustomize/v5/commands/version"
)

func TestInvalidOutputReturnsError(t *testing.T) {
	var out bytes.Buffer
	cmd := NewCmdVersion(&out)
	cmd.SetArgs([]string{"--output", "yml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid output error")
	}
	if err.Error() != "--output must be 'yaml' or 'json'" {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("expected no output, got %q", out.String())
	}
}
