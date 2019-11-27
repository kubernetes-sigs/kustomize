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
	resourceFileName    = "myWonderfulResource.yaml"
	resourceFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddResourceHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	fSys.WriteFile(resourceFileName+"another", []byte(resourceFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), resourceFileName) {
		t.Errorf("expected resource name in kustomization")
	}
	if !strings.Contains(string(content), resourceFileName+"another") {
		t.Errorf("expected resource name in kustomization")
	}
}

func TestAddResourceAlreadyThere(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(resourceFileName, []byte(resourceFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddResource(fSys)
	args := []string{resourceFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}

	// adding an existing resource doesn't return an error
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error :%v", err)
	}
}

func TestAddResourceNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddResource(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a resource file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
