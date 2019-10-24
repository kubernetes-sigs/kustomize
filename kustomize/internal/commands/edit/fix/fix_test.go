// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"strings"
	"testing"

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
