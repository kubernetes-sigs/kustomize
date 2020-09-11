// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

func TestFix(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, []byte(`nameprefix: some-prefix-`))

	cmd := NewCmdFix(fSys)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), "apiVersion: ") {
		t.Errorf("expected apiVersion in kustomization")
	}
	if !strings.Contains(string(content), "kind: Kustomization") {
		t.Errorf("expected kind in kustomization")
	}
}

func TestFixOutdatedPatchesFieldTitle(t *testing.T) {
	kustomizationContentWithOutdatedPatchesFieldTitle := []byte(`
patchesJson6902:
- path: patch1.yaml
  target:
    kind: Service
- path: patch2.yaml
  target:
    group: apps
    kind: Deployment
    version: v1
`)

	expected := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: patch1.yaml
  target:
    kind: Service
- path: patch2.yaml
  target:
    group: apps
    kind: Deployment
    version: v1
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomizationContentWithOutdatedPatchesFieldTitle)
	cmd := NewCmdFix(fSys)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), "apiVersion: ") {
		t.Errorf("expected apiVersion in kustomization")
	}
	if !strings.Contains(string(content), "kind: Kustomization") {
		t.Errorf("expected kind in kustomization")
	}

	if diff := cmp.Diff(expected, content); diff != "" {
		t.Errorf("Mismatch (-expected, +actual):\n%s", diff)
	}
}

func TestRenameAndKeepOutdatedPatchesField(t *testing.T) {
	kustomizationContentWithOutdatedPatchesFieldTitle := []byte(`
patchesJson6902:
- path: patch1.yaml
  target:
    kind: Deployment
patches:
- path: patch2.yaml
  target:
    kind: Deployment
- path: patch3.yaml
  target:
    kind: Service
`)

	expected := []byte(`
patches:
- path: patch2.yaml
  target:
    kind: Deployment
- path: patch3.yaml
  target:
    kind: Service
- path: patch1.yaml
  target:
    kind: Deployment
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomizationContentWithOutdatedPatchesFieldTitle)
	cmd := NewCmdFix(fSys)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), "apiVersion: ") {
		t.Errorf("expected apiVersion in kustomization")
	}
	if !strings.Contains(string(content), "kind: Kustomization") {
		t.Errorf("expected kind in kustomization")
	}

	if diff := cmp.Diff(expected, content); diff != "" {
		t.Errorf("Mismatch (-expected, +actual):\n%s", diff)
	}
}
