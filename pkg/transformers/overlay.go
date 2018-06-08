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

	jsonpatch "github.com/evanphx/json-patch"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
)

// overlayTransformer contains a map of overlay objects
type overlayTransformer struct {
	overlay []*resource.Resource
}

var _ Transformer = &overlayTransformer{}

// NewOverlayTransformer constructs a overlayTransformer.
func NewOverlayTransformer(overlay []*resource.Resource) (Transformer, error) {
	if len(overlay) == 0 {
		return NewNoOpTransformer(), nil
	}
	return &overlayTransformer{overlay}, nil
}

// Transform apply the overlay on top of the base resources.
func (o *overlayTransformer) Transform(baseResourceMap resmap.ResMap) error {
	// Merge and then index the patches by Id.
	overlays, err := o.mergePatches()
	if err != nil {
		return err
	}

	// Strategic merge the resources exist in both base and overlay.
	for _, overlay := range overlays {
		// Merge overlay with base resource.
		id := overlay.Id()
		base, found := baseResourceMap[id]
		if !found {
			return fmt.Errorf("failed to find an object with %#v to apply the patch", id.Gvk())
		}
		merged := map[string]interface{}{}
		versionedObj, err := scheme.Scheme.New(id.Gvk())
		baseName := base.GetName()
		switch {
		case runtime.IsNotRegisteredError(err):
			// Use JSON merge patch to handle types w/o schema
			baseBytes, err := json.Marshal(base)
			if err != nil {
				return err
			}
			patchBytes, err := json.Marshal(overlay)
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
			// Use Strategic Merge Patch to handle types w/ schema
			// TODO: Change this to use the new Merge package.
			// Store the name of the base object, because this name may have been munged.
			// Apply this name to the StrategicMergePatched object.
			lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObj)
			if err != nil {
				return err
			}
			merged, err = strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(
				base.Object,
				overlay.Object,
				lookupPatchMeta)
			if err != nil {
				return err
			}
		}
		base.SetName(baseName)
		baseResourceMap[id].Object = merged
	}
	return nil
}

// mergePatches merge and index patches by Id.
// It errors out if there is conflict between patches.
func (o *overlayTransformer) mergePatches() (resmap.ResMap, error) {
	rc := resmap.ResMap{}
	for ix, patch := range o.overlay {
		id := patch.Id()
		existing, found := rc[id]
		if !found {
			rc[id] = patch
			continue
		}

		versionedObj, err := scheme.Scheme.New(id.Gvk())
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
			conflictingPatch, err := cd.findConflict(ix, o.overlay)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("there is conflict between %#v and %#v", conflictingPatch.Object, patch.Object)
		}
		merged, err := cd.mergePatches(existing, patch)
		if err != nil {
			return nil, err
		}
		rc[id] = merged
	}
	return rc, nil
}
