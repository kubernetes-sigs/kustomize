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

package transformers

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

type conflictDetector interface {
	hasConflict(patch1, patch2 *unstructured.Unstructured) (bool, error)
	findConflict(conflictingPatchIdx int, patches []*unstructured.Unstructured) (*unstructured.Unstructured, error)
	mergePatches(patch1, patch2 *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type jsonMergePatch struct{}

var _ conflictDetector = &jsonMergePatch{}

func newJMPConflictDetector() conflictDetector {
	return &jsonMergePatch{}
}

func (jmp *jsonMergePatch) hasConflict(patch1, patch2 *unstructured.Unstructured) (bool, error) {
	return mergepatch.HasConflicts(patch1.Object, patch2.Object)
}

func (jmp *jsonMergePatch) findConflict(conflictingPatchIdx int, patches []*unstructured.Unstructured) (*unstructured.Unstructured, error) {
	for i, patch := range patches {
		if i == conflictingPatchIdx {
			continue
		}
		if patches[conflictingPatchIdx].GroupVersionKind() != patch.GroupVersionKind() ||
			patches[conflictingPatchIdx].GetName() != patch.GetName() {
			continue
		}
		conflict, err := mergepatch.HasConflicts(patch.Object, patches[conflictingPatchIdx].Object)
		if err != nil {
			return nil, err
		}
		if conflict {
			return patch, nil
		}
	}
	return nil, nil
}

func (jmp *jsonMergePatch) mergePatches(patch1, patch2 *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var merged unstructured.Unstructured
	var mergedMap map[string]interface{}
	baseBytes, err := json.Marshal(patch1.Object)
	if err != nil {
		return nil, err
	}
	patchBytes, err := json.Marshal(patch2.Object)
	if err != nil {
		return nil, err
	}
	mergedBytes, err := jsonpatch.MergeMergePatches(baseBytes, patchBytes)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(mergedBytes, &mergedMap)
	merged.SetUnstructuredContent(mergedMap)
	return &merged, err
}

type strategicMergePatch struct {
	lookupPatchMeta strategicpatch.LookupPatchMeta
}

var _ conflictDetector = &strategicMergePatch{}

func newSMPConflictDetector(versionedObj runtime.Object) (conflictDetector, error) {
	lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObj)
	return &strategicMergePatch{lookupPatchMeta: lookupPatchMeta}, err
}

func (smp *strategicMergePatch) hasConflict(patch1, patch2 *unstructured.Unstructured) (bool, error) {
	return strategicpatch.MergingMapsHaveConflicts(patch1.Object, patch2.Object, smp.lookupPatchMeta)
}

func (smp *strategicMergePatch) findConflict(conflictingPatchIdx int, patches []*unstructured.Unstructured) (*unstructured.Unstructured, error) {
	for i, patch := range patches {
		if i == conflictingPatchIdx {
			continue
		}
		if patches[conflictingPatchIdx].GroupVersionKind() != patch.GroupVersionKind() ||
			patches[conflictingPatchIdx].GetName() != patch.GetName() {
			continue
		}
		conflict, err := strategicpatch.MergingMapsHaveConflicts(
			patch.Object, patches[conflictingPatchIdx].Object, smp.lookupPatchMeta)
		if err != nil {
			return nil, err
		}
		if conflict {
			return patch, nil
		}
	}
	return nil, nil
}

func (smp *strategicMergePatch) mergePatches(patch1, patch2 *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	merged := unstructured.Unstructured{}
	mergeJsonMap, err := strategicpatch.MergeStrategicMergeMapPatchUsingLookupPatchMeta(
		smp.lookupPatchMeta, patch1.Object, patch2.Object)
	merged.SetUnstructuredContent(mergeJsonMap)
	return &merged, err
}
