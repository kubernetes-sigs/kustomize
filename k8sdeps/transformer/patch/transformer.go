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
	"sigs.k8s.io/kustomize/pkg/resid"
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
// nolint:ineffassign
func (tf *transformer) Transform(m resmap.ResMap) error {
	patches, err := tf.mergePatches()
	if err != nil {
		return err
	}
	for _, patch := range patches.Resources() {
		target, err := tf.findPatchTarget(m, patch.OrgId())
		merged := map[string]interface{}{}
		versionedObj, err := scheme.Scheme.New(
			toSchemaGvk(patch.OrgId().Gvk))
		saveName := target.GetName()
		switch {
		case runtime.IsNotRegisteredError(err):
			// Use JSON merge patch to handle types w/o schema
			baseBytes, err := json.Marshal(target.Map())
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
			// Store the name of the target object, because this name may have been munged.
			// Apply this name to the patched object.
			lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObj)
			if err != nil {
				return err
			}
			merged, err = strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(
				target.Map(),
				patch.Map(),
				lookupPatchMeta)
			if err != nil {
				return err
			}
		}
		target.SetMap(merged)
		target.SetName(saveName)
	}
	return nil
}

func (tf *transformer) findPatchTarget(
	m resmap.ResMap, id resid.ResId) (*resource.Resource, error) {
	match, err := m.GetByOriginalId(id)
	if err == nil {
		return match, nil
	}
	match, err = m.GetByCurrentId(id)
	if err == nil {
		return match, nil
	}
	return nil, fmt.Errorf(
		"failed to find target for patch %s", id.GvknString())
}

// mergePatches merge and index patches by OrgId.
// It errors out if there is conflict between patches.
func (tf *transformer) mergePatches() (resmap.ResMap, error) {
	rc := resmap.New()
	for ix, patch := range tf.patches {
		id := patch.OrgId()
		existing := rc.GetMatchingResourcesByOriginalId(id.GvknEquals)
		if len(existing) == 0 {
			rc.Append(patch)
			continue
		}
		if len(existing) > 1 {
			return nil, fmt.Errorf("self conflict in patches")
		}

		versionedObj, err := scheme.Scheme.New(toSchemaGvk(id.Gvk))
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

		conflict, err := cd.hasConflict(existing[0], patch)
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
		merged, err := cd.mergePatches(existing[0], patch)
		if err != nil {
			return nil, err
		}
		rc.Replace(merged)
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
