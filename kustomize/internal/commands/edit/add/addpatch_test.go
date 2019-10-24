// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

const (
	patchFileName    = "myWonderfulPatch.yaml"
	patchFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddPatchHappyPath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(patchFileName, []byte(patchFileContent))
	fSys.WriteFile(patchFileName+"another", []byte(patchFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{patchFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), patchFileName) {
		t.Errorf("expected patch name in kustomization")
	}
	if !strings.Contains(string(content), patchFileName+"another") {
		t.Errorf("expected patch name in kustomization")
	}
}

func TestAddPatchAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(patchFileName, []byte(patchFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{patchFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}

	// adding an existing patch shouldn't return an error
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
}

func TestAddPatchNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddPatch(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a patch file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
