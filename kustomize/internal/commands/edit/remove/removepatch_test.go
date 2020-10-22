// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

const (
	patchFileContent = `- op: replace
  path: /some/existing/path
  value: new value`
	kind               = "myKind"
	group              = "myGroup"
	version            = "myVersion"
	name               = "myName"
	namespace          = "myNamespace"
	annotationSelector = "myAnnotationSelector"
	labelSelector      = "myLabelSelector"
)

func makeKustomizationPatchFS() filesys.FileSystem {
	fSys := filesys.MakeEmptyDirInMemory()
	patches := []string{"patch1.yaml", "patch2.yaml"}

	testutils_test.WriteTestKustomizationWith(fSys, []byte(`
patches:
- path: patch1.yaml
  target:
    group: myGroup
    version: myVersion
    kind: myKind
    name: myName
    namespace: myNamespace
    labelSelector: myLabelSelector
    annotationSelector: myAnnotationSelector
- path: patch2.yaml
  target:
    group: myGroup
    version: myVersion
    kind: myKind
- patch: |-
    - op: replace
      path: /some/existing/path
      value: new value
  target:
    kind: myKind
    labelSelector: myLabelSelector
`))

	for _, p := range patches {
		fSys.WriteFile(p, []byte(patchFileContent))
	}
	fSys.WriteFile("patch3.yaml", []byte(patchFileContent))
	return fSys
}

func TestRemovePatch(t *testing.T) {
	fSys := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fSys)
	patchPath := "patch1.yaml"
	args := []string{
		"--path", patchPath,
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
		t.Fatalf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fSys)
	for _, p := range m.Patches {
		if p.Path == patchPath {
			t.Fatalf("%s must be deleted", patchPath)
		}
	}
}

func TestRemovePatch2(t *testing.T) {
	fSys := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fSys)
	args := []string{
		"--patch", patchFileContent,
		"--kind", kind,
		"--label-selector", labelSelector,
	}
	cmd.SetArgs(args)
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fSys)
	for _, p := range m.Patches {
		if p.Patch == patchFileContent {
			t.Fatalf("%s must be deleted", patchFileContent)
		}
	}
}

func TestRemovePatchNotDefinedInKustomization(t *testing.T) {
	fSys := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fSys)
	args := []string{
		"--path", "patch3.yaml",
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
		t.Fatalf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fSys)
	for _, k := range []string{"patch1.yaml", "patch2.yaml"} {
		found := false
		for _, p := range m.Patches {
			if p.Path == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s must exist", k)
		}
	}
}

func TestRemovePatchNotExist(t *testing.T) {
	fSys := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fSys)
	args := []string{
		"--path", "patch4.yaml",
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
		t.Fatalf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fSys)
	for _, k := range []string{"patch1.yaml", "patch2.yaml"} {
		found := false
		for _, p := range m.Patches {
			if p.Path == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s must exist", k)
		}
	}
}

func TestRemovePatchNoArgs(t *testing.T) {
	// if no flags specified, we should do nothing
	fSys := makeKustomizationPatchFS()
	cmd := newCmdRemovePatch(fSys)
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}
