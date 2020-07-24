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
	componentFileName    = "myWonderfulComponent.yaml"
	componentFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddComponentHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(componentFileName, []byte(componentFileContent))
	fSys.WriteFile(componentFileName+"another", []byte(componentFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddComponent(fSys)
	args := []string{componentFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), componentFileName) {
		t.Errorf("expected component name in kustomization")
	}
	if !strings.Contains(string(content), componentFileName+"another") {
		t.Errorf("expected component name in kustomization")
	}
}

func TestAddComponentAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(componentFileName, []byte(componentFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddComponent(fSys)
	args := []string{componentFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}

	// adding an existing component doesn't return an error
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error :%v", err)
	}
}

func TestAddComponentNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddComponent(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a component file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
