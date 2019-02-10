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
	"fmt"

	"github.com/evanphx/json-patch"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

// patchJson6902JSONTransformer applies patches.
type patchJson6902JSONTransformer struct {
	target resid.ResId
	patch  jsonpatch.Patch
}

var _ Transformer = &patchJson6902JSONTransformer{}

// newPatchJson6902JSONTransformer constructs a PatchJson6902 transformer.
func newPatchJson6902JSONTransformer(t resid.ResId, p jsonpatch.Patch) (Transformer, error) {
	if len(p) == 0 {
		return NewNoOpTransformer(), nil
	}
	return &patchJson6902JSONTransformer{target: t, patch: p}, nil
}

// Transform apply the json patches on top of the base resources.
func (t *patchJson6902JSONTransformer) Transform(baseResourceMap resmap.ResMap) error {
	obj, err := findTargetObj(baseResourceMap, t.target)
	if obj == nil {
		return err
	}
	rawObj, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	modifiedObj, err := t.patch.Apply(rawObj)
	if err != nil {
		return err
	}
	err = obj.UnmarshalJSON(modifiedObj)
	if err != nil {
		return err
	}
	return nil
}

func findTargetObj(m resmap.ResMap, targetId resid.ResId) (*resource.Resource, error) {
	matchedIds := m.FindByGVKN(targetId)
	if targetId.Namespace() != "" {
		var ids []resid.ResId
		for _, id := range matchedIds {
			if id.Namespace() == targetId.Namespace() {
				ids = append(ids, id)
			}
		}
		matchedIds = ids
	}

	if len(matchedIds) == 0 {
		return nil, fmt.Errorf("couldn't find any object to apply the json patch %v", targetId)
	}
	if len(matchedIds) > 1 {
		return nil, fmt.Errorf("found multiple objects that the patch can apply %v", matchedIds)
	}
	return m[matchedIds[0]], nil
}
