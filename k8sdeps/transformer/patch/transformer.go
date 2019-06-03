// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"encoding/json"
	"fmt"

	"github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

// transformer applies strategic merge patches.
type transformer struct {
	patches []*resource.Resource
	rf      *resource.Factory
}

var _ transformers.Transformer = &transformer{}

// NewTransformer constructs a strategic merge patch transformer.
func NewTransformer(
	slice []*resource.Resource, rf *resource.Factory) (transformers.Transformer, error) {
	if len(slice) == 0 {
		return transformers.NewNoOpTransformer(), nil
	}
	return &transformer{patches: slice, rf: rf}, nil
}

// Transform apply the patches on top of the base resources.
func (tf *transformer) Transform(baseResourceMap resmap.ResMap) error {
	// Merge and then index the patches by Id.
	patches, err := tf.mergePatches()
	if err != nil {
		return err
	}

	// Strategic merge the resources exist in both base and patches.
	for _, patch := range patches.Resources() {
		// Merge patches with base resource.
		id := patch.Id()
		matchedIds := baseResourceMap.GetMatchingIds(id.GvknEquals)
		if len(matchedIds) == 0 {
			return fmt.Errorf("failed to find an object with %s to apply the patch", id.GvknString())
		}
		if len(matchedIds) > 1 {
			return fmt.Errorf("found multiple objects %#v targeted by patch %#v (ambiguous)", matchedIds, id)
		}
		id = matchedIds[0]
		base := baseResourceMap.GetById(id)
		merged := map[string]interface{}{}
		versionedObj, err := scheme.Scheme.New(toSchemaGvk(id.Gvk()))
		baseName := base.GetName()
		switch {
		case runtime.IsNotRegisteredError(err):
			// Use JSON merge patch to handle types w/o schema
			baseBytes, err := json.Marshal(base.Map())
			if err != nil {
				return err
			}
			patchBytes, err := json.Marshal(patch.Map())
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
				base.Map(),
				patch.Map(),
				lookupPatchMeta)
			if err != nil {
				return err
			}
		}
		baseResourceMap.GetById(id).SetMap(merged)
		base.SetName(baseName)
	}
	return nil
}

// mergePatches merge and index patches by Id.
// It errors out if there is conflict between patches.
func (tf *transformer) mergePatches() (resmap.ResMap, error) {
	rc := resmap.New()
	for ix, patch := range tf.patches {
		id := patch.Id()
		existing := rc.GetById(id)
		if existing == nil {
			rc.AppendWithId(id, patch)
			continue
		}

		versionedObj, err := scheme.Scheme.New(toSchemaGvk(id.Gvk()))
		if err != nil && !runtime.IsNotRegisteredError(err) {
			return nil, err
		}
		var cd conflictDetector
		if err != nil {
			cd = newJMPConflictDetector(tf.rf)
		} else {
			cd, err = newSMPConflictDetector(versionedObj, tf.rf)
			if err != nil {
				return nil, err
			}
		}

		conflict, err := cd.hasConflict(existing, patch)
		if err != nil {
			return nil, err
		}
		if conflict {
			conflictingPatch, err := cd.findConflict(ix, tf.patches)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf(
				"conflict between %#v and %#v",
				conflictingPatch.Map(), patch.Map())
		}
		merged, err := cd.mergePatches(existing, patch)
		if err != nil {
			return nil, err
		}
		rc.ReplaceResource(id, merged)
	}
	return rc, nil
}

// toSchemaGvk converts to a schema.GroupVersionKind.
func toSchemaGvk(x gvk.Gvk) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   x.Group,
		Version: x.Version,
		Kind:    x.Kind,
	}
}
