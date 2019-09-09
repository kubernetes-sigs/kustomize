/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package patch

import "sigs.k8s.io/kustomize/v3/pkg/types"

// Append appends a slice of patch paths to a PatchStrategicMerge slice
func Append(patches []types.Patch, paths ...string) []types.Patch {
	for _, p := range paths {
		patches = append(patches, types.Patch{Path: p})
	}
	return patches
}

// Exist determines if a patch path exists in a slice of Patches
func Exist(patches []types.Patch, path string) bool {
	for _, p := range patches {
		if p.Path == path {
			return true
		}
	}
	return false
}

// Delete deletes patches from a Patch slice
func Delete(patches []types.Patch, paths ...string) []types.Patch {
	// Convert paths into Patch slice
	convertedPath := make([]types.Patch, len(paths))
	for i, p := range paths {
		convertedPath[i] = types.Patch{Path: p}
	}

	filteredPatches := make([]types.Patch, 0, len(patches))
	for _, containedPatch := range patches {
		if !Exist(convertedPath, containedPatch.Path) {
			filteredPatches = append(filteredPatches, containedPatch)
		}
	}
	return filteredPatches
}
