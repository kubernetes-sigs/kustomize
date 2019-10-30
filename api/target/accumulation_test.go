// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/api/target"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestTargetMustHaveKustomizationFile(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")
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
	_, err := th.MakeKustTargetOrErr()
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !IsMissingKustomizationFileError(err) {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestResourceDirectoryMustHaveKustomizationFile(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")
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
	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !IsMissingKustomizationFileError(err) {
		t.Fatalf("unexpected error: %q", err)
	}
}
