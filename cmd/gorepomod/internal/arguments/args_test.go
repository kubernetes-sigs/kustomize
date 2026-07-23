// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package arguments

import (
	"os"
	"testing"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

func TestParseReleaseBranchFlag(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	os.Args = []string{
		"gorepomod",
		"release",
		"kustomize",
		"minor",
		"--release-branch",
		"release-v5.8.2",
	}

	args, err := Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := args.ReleaseBranch(), "release-v5.8.2"; got != want {
		t.Fatalf("ReleaseBranch() = %q, want %q", got, want)
	}
	if got, want := args.Bump(), semver.Minor; got != want {
		t.Fatalf("Bump() = %v, want %v", got, want)
	}
}
