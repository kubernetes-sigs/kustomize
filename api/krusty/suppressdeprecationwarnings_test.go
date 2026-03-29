// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSuppressDeprecationWarnings(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
`)

	// Use the deprecated 'commonLabels' field to trigger a deprecation warning.
	th.WriteK(".", `
commonLabels:
  app: test
resources:
- deployment.yaml
`)

	// First, verify that warnings ARE produced without suppression.
	stderrWithWarnings := captureStderr(t, func() {
		opts := th.MakeDefaultOptions()
		th.Run(".", opts)
	})
	if !strings.Contains(stderrWithWarnings, "commonLabels") {
		t.Fatalf("expected deprecation warning about 'commonLabels', got: %q", stderrWithWarnings)
	}

	// Now, verify that warnings are NOT produced with suppression.
	stderrSuppressed := captureStderr(t, func() {
		opts := th.MakeDefaultOptions()
		opts.SuppressDeprecationWarnings = true
		th.Run(".", opts)
	})
	if strings.Contains(stderrSuppressed, "commonLabels") {
		t.Fatalf("expected no deprecation warning about 'commonLabels' when suppressed, got: %q", stderrSuppressed)
	}
}

func TestSuppressDeprecationWarningsWithVars(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  template:
    spec:
      containers:
      - name: app
        image: nginx
        env:
        - name: SERVICE_NAME
          value: $(SERVICE_NAME)
`)

	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: the-service
spec:
  ports:
  - port: 80
`)

	// Use the deprecated 'vars' field to trigger a deprecation warning.
	th.WriteK(".", `
resources:
- deployment.yaml
- service.yaml
vars:
- name: SERVICE_NAME
  objref:
    kind: Service
    name: the-service
    apiVersion: v1
`)

	// Verify that warnings are NOT produced with suppression.
	stderrSuppressed := captureStderr(t, func() {
		opts := th.MakeDefaultOptions()
		opts.SuppressDeprecationWarnings = true
		th.Run(".", opts)
	})
	if strings.Contains(stderrSuppressed, "vars") {
		t.Fatalf("expected no deprecation warning about 'vars' when suppressed, got: %q", stderrSuppressed)
	}
}

func TestSuppressDeprecationWarningsAPI(t *testing.T) {
	// Verify the API field exists and defaults to false.
	opts := krusty.MakeDefaultOptions()
	if opts.SuppressDeprecationWarnings {
		t.Fatal("expected SuppressDeprecationWarnings to default to false")
	}
}

// captureStderr captures anything written to os.Stderr during fn().
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	fn()

	w.Close()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}
