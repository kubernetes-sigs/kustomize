// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

func TestFix(t *testing.T) {
	fSys := fs.MakeFsInMemory()
	testutils.WriteTestKustomizationWith(fSys, []byte(`nameprefix: some-prefix-`))

	cmd := NewCmdFix(fSys)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils.ReadTestKustomization(fSys)
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
