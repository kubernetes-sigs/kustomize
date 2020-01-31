// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"path/filepath"
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestTargetMustHaveKustomizationFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: aService
`)
	th.WriteF("/app/deeper/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: anotherService
`)
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !IsMissingKustomizationFileError(err) {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestTargetMustHaveOnlyOneKustomizationFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		th.WriteF(filepath.Join("/app", n), `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`)
	}
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "Found multiple kustomization files under: /app") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestBaseMustHaveKustomizationFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
resources:
- base
`)
	th.WriteF("/app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    backend: bungie
  ports:
    - port: 7002
`)
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "accumulating resources") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestResourceNotFound(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
resources:
- deployment.yaml
`)
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "accumulating resources") {
		t.Fatalf("unexpected error: %q", err)
	}
}
