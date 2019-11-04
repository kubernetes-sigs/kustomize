// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patch

import "sigs.k8s.io/kustomize/api/types"

// Append appends a slice of patch paths to a PatchStrategicMerge slice
func Append(patches []types.PatchStrategicMerge, paths ...string) []types.PatchStrategicMerge {
	for _, p := range paths {
		patches = append(patches, types.PatchStrategicMerge(p))
	}
	return patches
}

// Exist determines if a patch path exists in a slice of PatchStrategicMerge
func Exist(patches []types.PatchStrategicMerge, path string) bool {
	for _, p := range patches {
		if p == types.PatchStrategicMerge(path) {
			return true
		}
	}
	return false
}

// Delete deletes patches from a PatchStrategicMerge slice
func Delete(patches []types.PatchStrategicMerge, paths ...string) []types.PatchStrategicMerge {
	// Convert paths into PatchStrategicMerge slice
	convertedPath := make([]types.PatchStrategicMerge, len(paths))
	for i, p := range paths {
		convertedPath[i] = types.PatchStrategicMerge(p)
	}

	filteredPatches := make([]types.PatchStrategicMerge, 0, len(patches))
	for _, containedPatch := range patches {
		if !Exist(convertedPath, string(containedPatch)) {
			filteredPatches = append(filteredPatches, containedPatch)
		}
	}
	return filteredPatches
}
