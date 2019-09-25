// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	resourceFileName    = "myWonderfulResource.yaml"
	resourceFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddResourceHappyPath(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(resourceFileName, []byte(resourceFileContent))
	fakeFS.WriteFile(resourceFileName+"another", []byte(resourceFileContent))
	fakeFS.WriteTestKustomization()

	cmd := newCmdAddResource(fakeFS)
	args := []string{resourceFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadTestKustomization()
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
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(resourceFileName, []byte(resourceFileContent))
	fakeFS.WriteTestKustomization()

	cmd := newCmdAddResource(fakeFS)
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
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdAddResource(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a resource file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
