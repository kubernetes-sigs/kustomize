// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

const (
	goodNamespaceValue = "staging"
)

func TestSetNamespaceHappyPath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdSetNamespace(fSys, valtest_test.MakeFakeValidator())
	args := []string{goodNamespaceValue}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	expected := []byte(fmt.Sprintf("namespace: %s", goodNamespaceValue))
	if !strings.Contains(string(content), string(expected)) {
		t.Errorf("expected namespace in kustomization file")
	}
}

func TestSetNamespaceOverride(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdSetNamespace(fSys, valtest_test.MakeFakeValidator())
	args := []string{goodNamespaceValue}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	args = []string{"newnamespace"}
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	expected := []byte("namespace: newnamespace")
	if !strings.Contains(string(content), string(expected)) {
		t.Errorf("expected namespace in kustomization file %s", string(content))
	}
}

func TestSetNamespaceNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdSetNamespace(fSys, valtest_test.MakeFakeValidator())
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify exactly one namespace value" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestSetNamespaceInvalid(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdSetNamespace(fSys, valtest_test.MakeFakeValidator())
	args := []string{"/badnamespace/"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if !strings.Contains(err.Error(), "is not a valid namespace name") {
		t.Errorf("unexpected error: %v", err.Error())
	}
}
