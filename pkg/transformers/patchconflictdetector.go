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
	"github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/kustomize/pkg/resource"
)

type conflictDetector interface {
	hasConflict(patch1, patch2 *resource.Resource) (bool, error)
	findConflict(conflictingPatchIdx int, patches []*resource.Resource) (*resource.Resource, error)
	mergePatches(patch1, patch2 *resource.Resource) (*resource.Resource, error)
}

type jsonMergePatch struct{}

var _ conflictDetector = &jsonMergePatch{}

func newJMPConflictDetector() conflictDetector {
	return &jsonMergePatch{}
}

func (jmp *jsonMergePatch) hasConflict(patch1, patch2 *resource.Resource) (bool, error) {
	return mergepatch.HasConflicts(patch1.FunStruct().Map(), patch2.FunStruct().Map())
}

func (jmp *jsonMergePatch) findConflict(conflictingPatchIdx int, patches []*resource.Resource) (*resource.Resource, error) {
	for i, patch := range patches {
		if i == conflictingPatchIdx {
			continue
		}
		if !patches[conflictingPatchIdx].FunStruct().GetGvk().Equals(patch.FunStruct().GetGvk()) ||
			patches[conflictingPatchIdx].FunStruct().GetName() != patch.FunStruct().GetName() {
			continue
		}
		conflict, err := mergepatch.HasConflicts(
			patch.FunStruct().Map(),
			patches[conflictingPatchIdx].FunStruct().Map())
		if err != nil {
			return nil, err
		}
		if conflict {
			return patch, nil
		}
	}
	return nil, nil
}

func (jmp *jsonMergePatch) mergePatches(patch1, patch2 *resource.Resource) (*resource.Resource, error) {
	baseBytes, err := json.Marshal(patch1.FunStruct().Map())
	if err != nil {
		return nil, err
	}
	patchBytes, err := json.Marshal(patch2.FunStruct().Map())
	if err != nil {
		return nil, err
	}
	mergedBytes, err := jsonpatch.MergeMergePatches(baseBytes, patchBytes)
	if err != nil {
		return nil, err
	}
	mergedMap := make(map[string]interface{})
	err = json.Unmarshal(mergedBytes, &mergedMap)
	return resource.NewResourceFromMap(mergedMap), err
}

type strategicMergePatch struct {
	lookupPatchMeta strategicpatch.LookupPatchMeta
}

var _ conflictDetector = &strategicMergePatch{}

func newSMPConflictDetector(versionedObj runtime.Object) (conflictDetector, error) {
	lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObj)
	return &strategicMergePatch{lookupPatchMeta: lookupPatchMeta}, err
}

func (smp *strategicMergePatch) hasConflict(patch1, patch2 *resource.Resource) (bool, error) {
	return strategicpatch.MergingMapsHaveConflicts(
		patch1.FunStruct().Map(),
		patch2.FunStruct().Map(), smp.lookupPatchMeta)
}

func (smp *strategicMergePatch) findConflict(conflictingPatchIdx int, patches []*resource.Resource) (*resource.Resource, error) {
	for i, patch := range patches {
		if i == conflictingPatchIdx {
			continue
		}
		if !patches[conflictingPatchIdx].FunStruct().GetGvk().Equals(patch.FunStruct().GetGvk()) ||
			patches[conflictingPatchIdx].FunStruct().GetName() != patch.FunStruct().GetName() {
			continue
		}
		conflict, err := strategicpatch.MergingMapsHaveConflicts(
			patch.FunStruct().Map(),
			patches[conflictingPatchIdx].FunStruct().Map(),
			smp.lookupPatchMeta)
		if err != nil {
			return nil, err
		}
		if conflict {
			return patch, nil
		}
	}
	return nil, nil
}

func (smp *strategicMergePatch) mergePatches(patch1, patch2 *resource.Resource) (*resource.Resource, error) {
	mergeJsonMap, err := strategicpatch.MergeStrategicMergeMapPatchUsingLookupPatchMeta(
		smp.lookupPatchMeta, patch1.FunStruct().Map(), patch2.FunStruct().Map())
	return resource.NewResourceFromMap(mergeJsonMap), err
}
