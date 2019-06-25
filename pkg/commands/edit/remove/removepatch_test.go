// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/patch"
)

const (
	patchFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func makeKustomizationPatchFS() fs.FileSystem {
	fakeFS := fs.MakeFakeFS()
	patches := []string{"patch1.yaml", "patch2.yaml"}

	fakeFS.WriteTestKustomizationWith([]byte(
		fmt.Sprintf("patchesStrategicMerge:\n  - %s",
			strings.Join(patches, "\n  - "))))

	for _, p := range patches {
		fakeFS.WriteFile(p, []byte(patchFileContent))
	}
	fakeFS.WriteFile("patch3.yaml", []byte(patchFileContent))
	return fakeFS
}

func TestRemovePatch(t *testing.T) {
	fakeFS := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fakeFS)
	args := []string{"patch1.yaml"}
	err := cmd.RunE(cmd, args)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	for _, k := range args {
		if patch.Exist(m.PatchesStrategicMerge, k) {
			t.Errorf("%s must be deleted", k)
		}
	}
}

func TestRemovePatchMultipleArgs(t *testing.T) {
	fakeFS := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fakeFS)
	args := []string{"patch1.yaml", "patch2.yaml"}
	err := cmd.RunE(cmd, args)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	for _, k := range args {
		if patch.Exist(m.PatchesStrategicMerge, k) {
			t.Errorf("%s must be deleted", k)
		}
	}
}

func TestRemovePatchGlob(t *testing.T) {
	fakeFS := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fakeFS)
	args := []string{"patch*.yaml"}
	err := cmd.RunE(cmd, args)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	if len(m.PatchesStrategicMerge) != 0 {
		t.Errorf("all patch must be deleted")
	}
}

func TestRemovePatchNotDefinedInKustomization(t *testing.T) {
	fakeFS := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fakeFS)
	args := []string{"patch3.yaml"}
	err := cmd.RunE(cmd, args)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	for _, k := range []string{"patch1.yaml", "patch2.yaml"} {
		if !patch.Exist(m.PatchesStrategicMerge, k) {
			t.Errorf("%s must exist", k)
		}
	}
}

func TestRemovePatchNotExist(t *testing.T) {
	fakeFS := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fakeFS)
	args := []string{"patch4.yaml"}
	err := cmd.RunE(cmd, args)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	for _, k := range []string{"patch1.yaml", "patch2.yaml"} {
		if !patch.Exist(m.PatchesStrategicMerge, k) {
			t.Errorf("%s must exist", k)
		}
	}
}

func TestRemovePatchNoArgs(t *testing.T) {
	fakeFS := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fakeFS)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify a patch file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
