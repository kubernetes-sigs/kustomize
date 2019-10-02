// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	baseDirectoryPaths = "my/path/to/wonderful/base,other/path/to/even/more/wonderful/base"
)

func TestAddBaseHappyPath(t *testing.T) {
	fSys := fs.MakeFsInMemory()
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		fSys.Mkdir(base)
	}
	testutils.WriteTestKustomization(fSys)

	cmd := newCmdAddBase(fSys)
	args := []string{baseDirectoryPaths}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}

	for _, base := range bases {
		if !strings.Contains(string(content), base) {
			t.Errorf("expected base name in kustomization")
		}
	}
}

func TestAddBaseAlreadyThere(t *testing.T) {
	fSys := fs.MakeFsInMemory()
	// Create fake directories
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		fSys.Mkdir(base)
	}
	testutils.WriteTestKustomization(fSys)

	cmd := newCmdAddBase(fSys)
	args := []string{baseDirectoryPaths}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}
	// adding an existing base should return an error
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected already there problem")
	}
	var expectedErrors []string
	for _, base := range bases {
		msg := "base " + base + " already in kustomization file"
		expectedErrors = append(expectedErrors, msg)
		if !kustfile.StringInSlice(msg, expectedErrors) {
			t.Errorf("unexpected error %v", err)
		}
	}
}

func TestAddBaseNoArgs(t *testing.T) {
	fSys := fs.MakeFsInMemory()

	cmd := newCmdAddBase(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a base directory" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
