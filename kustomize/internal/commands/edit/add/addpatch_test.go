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
`
	kind               = "myKind"
	group              = "myGroup"
	version            = "myVersion"
	name               = "myName"
	namespace          = "myNamespace"
	annotationSelector = "myAnnotationSelector"
	labelSelector      = "myLabelSelector"
)

func TestAddPatchWithFilePath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(patchFileName, []byte(patchFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{
		"--path", patchFileName,
		"--kind", kind,
		"--group", group,
		"--version", version,
		"--name", name,
		"--namespace", namespace,
		"--annotation-selector", annotationSelector,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	for i := 1; i < len(args); i += 2 {
		if !strings.Contains(string(content), args[i]) {
			t.Errorf("expected flag value of %s in kustomization but got\n%s", args[i-1], content)
		}
	}
}

func TestAddPatchWithPatchContent(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(patchFileName, []byte(patchFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{
		"--patch", patchFileContent,
		"--kind", kind,
		"--group", group,
		"--version", version,
		"--name", name,
		"--namespace", namespace,
		"--annotation-selector", annotationSelector,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	for i := 1; i < len(args); i += 2 {
		if !strings.Contains(string(content), strings.Trim(args[i], " \n")) {
			t.Errorf("expected flag value of %s in kustomization but got\n%s", args[i-1], content)
		}
	}
}

func TestAddPatchAlreadyThere(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(patchFileName, []byte(patchFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddPatch(fSys)
	args := []string{
		"--path", patchFileName,
		"--kind", kind,
		"--group", group,
		"--version", version,
		"--name", name,
		"--namespace", namespace,
		"--annotation-selector", annotationSelector,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}

	// adding an existing patch shouldn't return an error
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
}

func TestAddPatchNoArgs(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()

	cmd := newCmdAddPatch(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must provide either patch or path" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
