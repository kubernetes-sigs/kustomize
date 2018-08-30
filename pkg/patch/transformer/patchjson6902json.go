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
	"fmt"
	"log"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubernetes-sigs/kustomize/pkg/patch"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
)

// patchJson6902Transformer applies patches.
type patchJson6902JSONTransformer struct {
	target     *patch.Target
	operations []byte
}

var _ transformers.Transformer = &patchJson6902JSONTransformer{}

// NewPatchJson6902JSONTransformer constructs a PatchJson6902 transformer.
func NewPatchJson6902JSONTransformer(t *patch.Target, o []byte) (transformers.Transformer, error) {
	return &patchJson6902JSONTransformer{target: t, operations: o}, nil
}

// Transform apply the json patches on top of the base resources.
func (t *patchJson6902JSONTransformer) Transform(baseResourceMap resmap.ResMap) error {
	targetId := resource.NewResIdWithPrefixNamespace(
		schema.GroupVersionKind{
			Group:   t.target.Group,
			Version: t.target.Version,
			Kind:    t.target.Kind,
		},
		t.target.Name,
		"",
		t.target.Namespace,
	)

	matchedIds := baseResourceMap.FindByGVKN(targetId)
	if targetId.Namespace() != "" {
		ids := []resource.ResId{}
		for _, id := range matchedIds {
			if id.Namespace() == targetId.Namespace() {
				ids = append(ids, id)
			}
		}
		matchedIds = ids
	}
	if len(matchedIds) == 0 {
		log.Printf("Couldn't find any object to apply the json patch %v, skipping it.", targetId)
		return nil
	}
	if len(matchedIds) > 1 {
		return fmt.Errorf("found multiple objects that the patch can apply %v", matchedIds)
	}

	decodedPatch, err := jsonpatch.DecodePatch(t.operations)
	if err != nil {
		return err
	}
	obj := baseResourceMap[matchedIds[0]]
	rawObj, err := obj.Unstructured.MarshalJSON()
	if err != nil {
		return err
	}
	modifiedObj, err := decodedPatch.Apply(rawObj)
	if err != nil {
		return err
	}
	err = obj.UnmarshalJSON(modifiedObj)
	if err != nil {
		return err
	}
	return nil
}
