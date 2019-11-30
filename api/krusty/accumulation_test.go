// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/api/internal/target"
)

func TestTargetMustHaveKustomizationFile(t *testing.T) {
	th := makeTestHarness(t)
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

func TestBaseMustHaveKustomizationFile(t *testing.T) {
	th := makeTestHarness(t)
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
	if !IsMissingKustomizationFileError(err) {
		t.Fatalf("unexpected error: %q", err)
	}
}
