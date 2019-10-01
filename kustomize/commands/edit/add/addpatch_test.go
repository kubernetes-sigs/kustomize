// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	patchFileName    = "myWonderfulPatch.yaml"
	patchFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddPatchHappyPath(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(patchFileName, []byte(patchFileContent))
	fakeFS.WriteFile(patchFileName+"another", []byte(patchFileContent))
	fakeFS.WriteTestKustomization()

	cmd := newCmdAddPatch(fakeFS)
	args := []string{patchFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadTestKustomization()
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
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(patchFileName, []byte(patchFileContent))
	fakeFS.WriteTestKustomization()

	cmd := newCmdAddPatch(fakeFS)
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
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdAddPatch(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a patch file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
