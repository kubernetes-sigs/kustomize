// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package conflict

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"sigs.k8s.io/kustomize/api/resource"
)

// conflictDetectorJson detects conflicts in a list of JSON patches.
type conflictDetectorJson struct {
	resourceFactory *resource.Factory
}

var _ resource.ConflictDetector = &conflictDetectorJson{}

func (cd *conflictDetectorJson) HasConflict(
	p1, p2 *resource.Resource) (bool, error) {
	return mergepatch.HasConflicts(p1.Map(), p2.Map())
}

func (cd *conflictDetectorJson) MergePatches(
	patch1, patch2 *resource.Resource) (*resource.Resource, error) {
	baseBytes, err := json.Marshal(patch1.Map())
	if err != nil {
		return nil, err
	}
	patchBytes, err := json.Marshal(patch2.Map())
	if err != nil {
		return nil, err
	}
	mergedBytes, err := jsonpatch.MergeMergePatches(baseBytes, patchBytes)
	if err != nil {
		return nil, err
	}
	mergedMap := make(map[string]interface{})
	err = json.Unmarshal(mergedBytes, &mergedMap)
	return cd.resourceFactory.FromMap(mergedMap), err
}
