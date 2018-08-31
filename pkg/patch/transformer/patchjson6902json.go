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

package transformer

import (
	"github.com/evanphx/json-patch"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
)

// patchJson6902Transformer applies patches.
type patchJson6902JSONTransformer struct {
	target resource.ResId
	patch  jsonpatch.Patch
}

var _ transformers.Transformer = &patchJson6902JSONTransformer{}

// newPatchJson6902JSONTransformer constructs a PatchJson6902 transformer.
func newPatchJson6902JSONTransformer(t resource.ResId, p jsonpatch.Patch) (transformers.Transformer, error) {
	if len(p) == 0 {
		return transformers.NewNoOpTransformer(), nil
	}
	return &patchJson6902JSONTransformer{target: t, patch: p}, nil
}

// Transform apply the json patches on top of the base resources.
func (t *patchJson6902JSONTransformer) Transform(baseResourceMap resmap.ResMap) error {
	obj, err := findTargetObj(baseResourceMap, t.target)
	if obj == nil {
		return err
	}
	rawObj, err := obj.Unstructured.MarshalJSON()
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
