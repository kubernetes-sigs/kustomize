// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// TestSuppressDeprecationWarnings verifies that SuppressDeprecationWarnings=true
// prevents deprecation messages from being written to stderr.
func TestSuppressDeprecationWarnings(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("dep.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: mycm
`)
	// Use deprecated 'bases' field to trigger a deprecation warning.
	th.WriteK(".", `
bases:
- dep.yaml
`)

	captureStderr := func(fn func()) string {
		r, w, _ := os.Pipe()
		old := os.Stderr
		os.Stderr = w
		fn()
		w.Close()
		os.Stderr = old
		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

	// With default options, deprecation warning should appear on stderr.
	withWarning := captureStderr(func() {
		opts := th.MakeDefaultOptions()
		th.Run(".", opts)
	})
	if !strings.Contains(withWarning, "deprecated") {
		t.Errorf("expected deprecation warning on stderr, got: %q", withWarning)
	}

	// With SuppressDeprecationWarnings=true, stderr should be silent.
	withoutWarning := captureStderr(func() {
		opts := th.MakeDefaultOptions()
		opts.SuppressDeprecationWarnings = true
		th.Run(".", opts)
	})
	if withoutWarning != "" {
		t.Errorf("expected no output on stderr when suppressed, got: %q", withoutWarning)
	}
}

// A simple usage example to shows what happens when
// there are no files to read.
// For more substantial tests and examples,
// see other tests in this package.
func TestEmptyFileSystem(t *testing.T) {
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	_, err := b.Run(filesys.MakeFsInMemory(), "noSuchThing")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "'noSuchThing' doesn't exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}
