// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.

package patch

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/types"
)

func buildPatchStrategicMergeSlice(patchStrings []string) []types.PatchStrategicMerge {
	var patches []types.PatchStrategicMerge
	for _, patchString := range patchStrings {
		patches = append(patches, types.PatchStrategicMerge(patchString))
	}
	return patches
}

func TestAppend(t *testing.T) {
	patchStrings := []string{"patch1.yaml", "patch2.yaml"}
	patches := buildPatchStrategicMergeSlice(patchStrings)

	patches = Append(patches, "patch3.yaml")

	for i, k := range []string{"patch1.yaml", "patch2.yaml", "patch3.yaml"} {
		if patches[i] != types.PatchStrategicMerge(k) {
			t.Fatalf("patches[%d] must be %s, got %s", i, k, patches[i])
		}
	}
}

func TestExistTrue(t *testing.T) {
	patchStrings := []string{"patch1.yaml", "patch2.yaml"}
	patches := buildPatchStrategicMergeSlice(patchStrings)

	for _, patchString := range patchStrings {
		if !Exist(patches, patchString) {
			t.Fatalf("%s must exist", patchString)
		}
	}
}

func TestExistFalse(t *testing.T) {
	patchStrings := []string{"patch1.yaml", "patch2.yaml"}
	patches := buildPatchStrategicMergeSlice(patchStrings)

	for _, patchString := range []string{"invalid1.yaml", "invalid2.yaml"} {
		if Exist(patches, patchString) {
			t.Fatalf("%s must not exist", patchString)
		}
	}
}

func TestDelete(t *testing.T) {
	patchStrings := []string{"patch1.yaml", "patch2.yaml"}
	patches := buildPatchStrategicMergeSlice(patchStrings)

	patches = Delete(patches, "patch1.yaml")

	if Exist(patches, "patch1.yaml") {
		t.Fatalf("patch1.yaml should be deleted")
	}
	if !Exist(patches, "patch2.yaml") {
		t.Fatalf("patch2.yaml should exist")
	}
	if len(patches) != 1 {
		t.Fatalf("Length of slice must be 1: actual %d", len(patches))
	}
}

func TestDeleteMultiple(t *testing.T) {
	patchStrings := []string{"patch1.yaml", "patch2.yaml"}
	patches := buildPatchStrategicMergeSlice(patchStrings)

	patches = Delete(patches, "patch2.yaml", "patch4.yaml", "patch1.yaml", "patch3.yaml")

	for _, k := range patchStrings {
		if Exist(patches, k) {
			t.Fatalf("%s should be deleted", k)
		}
	}
}
