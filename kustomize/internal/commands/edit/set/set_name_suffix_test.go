// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

const (
	goodSuffixValue = "-acme"
)

func TestSetNameSuffixHappyPath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdSetNameSuffix(fSys)
	args := []string{goodSuffixValue}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), goodSuffixValue) {
		t.Errorf("expected suffix value in kustomization file")
	}
}

func TestSetNameSuffixNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	cmd := newCmdSetNameSuffix(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify exactly one suffix value" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
