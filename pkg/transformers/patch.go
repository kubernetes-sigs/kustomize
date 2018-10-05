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
	"fmt"

	"github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

// patchTransformer applies patches.
type patchTransformer struct {
	patches []*resource.Resource
}

var _ Transformer = &patchTransformer{}

// NewPatchTransformer constructs a patchTransformer.
func NewPatchTransformer(slice []*resource.Resource) (Transformer, error) {
	if len(slice) == 0 {
		return NewNoOpTransformer(), nil
	}
	return &patchTransformer{slice}, nil
}

// Transform apply the patches on top of the base resources.
func (pt *patchTransformer) Transform(baseResourceMap resmap.ResMap) error {
	// Merge and then index the patches by Id.
	patches, err := pt.mergePatches()
	if err != nil {
		return err
	}

	// Strategic merge the resources exist in both base and patches.
	for _, patch := range patches {
		// Merge patches with base resource.
		id := patch.Id()
		matchedIds := baseResourceMap.FindByGVKN(id)
		if len(matchedIds) == 0 {
			return fmt.Errorf("failed to find an object with %#v to apply the patch", id.Gvk())
		}
		if len(matchedIds) > 1 {
			return fmt.Errorf("found multiple objects %#v targeted by patch %#v (ambiguous)", matchedIds, id)
		}
		id = matchedIds[0]
		base := baseResourceMap[id]
		merged := map[string]interface{}{}
		versionedObj, err := scheme.Scheme.New(id.Gvk().ToSchemaGvk())
		baseName := base.FunStruct().GetName()
		switch {
		case runtime.IsNotRegisteredError(err):
			// Use JSON merge patch to handle types w/o schema
			baseBytes, err := json.Marshal(base.FunStruct().Map())
			if err != nil {
				return err
			}
			patchBytes, err := json.Marshal(patch.FunStruct().Map())
			if err != nil {
				return err
			}
			mergedBytes, err := jsonpatch.MergePatch(baseBytes, patchBytes)
			if err != nil {
				return err
			}
			err = json.Unmarshal(mergedBytes, &merged)
			if err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			// Use Strategic-Merge-Patch to handle types w/ schema
			// TODO: Change this to use the new Merge package.
			// Store the name of the base object, because this name may have been munged.
			// Apply this name to the patched object.
			lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObj)
			if err != nil {
				return err
			}
			merged, err = strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(
				base.FunStruct().Map(),
				patch.FunStruct().Map(),
				lookupPatchMeta)
			if err != nil {
				return err
			}
		}
		base.FunStruct().SetName(baseName)
		baseResourceMap[id].FunStruct().SetMap(merged)
	}
	return nil
}

// mergePatches merge and index patches by Id.
// It errors out if there is conflict between patches.
func (pt *patchTransformer) mergePatches() (resmap.ResMap, error) {
	rc := resmap.ResMap{}
	for ix, patch := range pt.patches {
		id := patch.Id()
		existing, found := rc[id]
		if !found {
			rc[id] = patch
			continue
		}

		versionedObj, err := scheme.Scheme.New(id.Gvk().ToSchemaGvk())
		if err != nil && !runtime.IsNotRegisteredError(err) {
			return nil, err
		}
		var cd conflictDetector
		if err != nil {
			cd = newJMPConflictDetector()
		} else {
			cd, err = newSMPConflictDetector(versionedObj)
			if err != nil {
				return nil, err
			}
		}

		conflict, err := cd.hasConflict(existing, patch)
		if err != nil {
			return nil, err
		}
		if conflict {
			conflictingPatch, err := cd.findConflict(ix, pt.patches)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf(
				"conflict between %#v and %#v",
				conflictingPatch.FunStruct().Map(), patch.FunStruct().Map())
		}
		merged, err := cd.mergePatches(existing, patch)
		if err != nil {
			return nil, err
		}
		rc[id] = merged
	}
	return rc, nil
}
