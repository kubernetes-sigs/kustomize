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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/kustomize/resource"
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
func (o *overlayTransformer) Transform(baseResourceMap resource.ResourceCollection) error {
	// Merge and then index the patches by GVKN.
	overlays, err := o.mergePatches()
	if err != nil {
		return err
	}

	// Strategic merge the resources exist in both base and overlay.
	for _, overlay := range overlays {
		// Merge overlay with base resource.
		gvkn := overlay.GVKN()
		base, found := baseResourceMap[gvkn]
		if !found {
			return fmt.Errorf("failed to find an object with %#v to apply the patch", gvkn.GVK)
		}
		merged := map[string]interface{}{}
		versionedObj, err := scheme.Scheme.New(gvkn.GVK)
		baseName := base.Data.GetName()
		switch {
		case runtime.IsNotRegisteredError(err):
			// Use JSON merge patch to handle types w/o schema
			baseBytes, err := json.Marshal(base.Data)
			if err != nil {
				return err
			}
			patchBytes, err := json.Marshal(overlay.Data)
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
				base.Data.Object,
				overlay.Data.Object,
				lookupPatchMeta)
			if err != nil {
				return err
			}
		}
		base.Data.SetName(baseName)
		baseResourceMap[gvkn].Data.Object = merged
	}
	return nil
}

// mergePatches merge and index patches by GVKN.
// It errors out if there is conflict between patches.
func (o *overlayTransformer) mergePatches() (resource.ResourceCollection, error) {
	rc := resource.ResourceCollection{}
	patches := resourcesToObjects(o.overlay)
	for ix, patch := range o.overlay {
		gvkn := patch.GVKN()
		existing, found := rc[gvkn]
		if !found {
			rc[gvkn] = patch
			continue
		}

		versionedObj, err := scheme.Scheme.New(gvkn.GVK)
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

		conflict, err := cd.hasConflict(existing.Data, patch.Data)
		if err != nil {
			return nil, err
		}
		if conflict {
			conflictingPatch, err := cd.findConflict(ix, patches)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("there is conflict between %#v and %#v", conflictingPatch.Object, patch.Data.Object)
		} else {
			merged, err := cd.mergePatches(existing.Data, patch.Data)
			if err != nil {
				return nil, err
			}
			existing.Data = merged
		}
	}
	return rc, nil
}

func resourcesToObjects(rs []*resource.Resource) []*unstructured.Unstructured {
	objectList := make([]*unstructured.Unstructured, len(rs))
	for i := range rs {
		objectList[i] = rs[i].Data
	}
	return objectList
}
