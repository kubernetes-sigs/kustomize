// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package conflict

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/kustomize/api/resource"
)

// conflictDetectorSm detects conflicts in a list of strategic merge patches.
type conflictDetectorSm struct {
	lookupPatchMeta strategicpatch.LookupPatchMeta
	resourceFactory *resource.Factory
}

var _ resource.ConflictDetector = &conflictDetectorSm{}

func (cd *conflictDetectorSm) HasConflict(
	p1, p2 *resource.Resource) (bool, error) {
	return strategicpatch.MergingMapsHaveConflicts(
		p1.Map(), p2.Map(), cd.lookupPatchMeta)
}

func (cd *conflictDetectorSm) MergePatches(
	patch1, patch2 *resource.Resource) (*resource.Resource, error) {
	if cd.hasDeleteDirectiveMarker(patch2.Map()) {
		if cd.hasDeleteDirectiveMarker(patch1.Map()) {
			return nil, fmt.Errorf(
				"cannot merge patches both containing '$patch: delete' directives")
		}
		patch1, patch2 = patch2, patch1
	}
	mergedMap, err := strategicpatch.MergeStrategicMergeMapPatchUsingLookupPatchMeta(
		cd.lookupPatchMeta, patch1.Map(), patch2.Map())
	return cd.resourceFactory.FromMap(mergedMap), err
}

func (cd *conflictDetectorSm) hasDeleteDirectiveMarker(
	patch map[string]interface{}) bool {
	if v, ok := patch["$patch"]; ok && v == "delete" {
		return true
	}
	for _, v := range patch {
		switch typedV := v.(type) {
		case map[string]interface{}:
			if cd.hasDeleteDirectiveMarker(typedV) {
				return true
			}
		case []interface{}:
			for _, sv := range typedV {
				typedE, ok := sv.(map[string]interface{})
				if !ok {
					break
				}
				if cd.hasDeleteDirectiveMarker(typedE) {
					return true
				}
			}
		}
	}
	return false
}
