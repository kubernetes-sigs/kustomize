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

	"github.com/ghodss/yaml"
	yamlpatch "github.com/krishicks/yaml-patch"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubernetes-sigs/kustomize/pkg/patch"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
)

// patchJson6902Transformer applies patches.
type patchJson6902YAMLTransformer struct {
	target *patch.Target
	patch  yamlpatch.Patch
}

var _ transformers.Transformer = &patchJson6902YAMLTransformer{}

// NewPatchJson6902YAMLTransformer constructs a PatchJson6902 transformer.
func NewPatchJson6902YAMLTransformer(t *patch.Target, p yamlpatch.Patch) (transformers.Transformer, error) {
	return &patchJson6902YAMLTransformer{target: t, patch: p}, nil
}

// Transform apply the json patches on top of the base resources.
func (t *patchJson6902YAMLTransformer) Transform(baseResourceMap resmap.ResMap) error {
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

	obj := baseResourceMap[matchedIds[0]]
	rawObj, err := yaml.Marshal(obj.Unstructured.Object)
	if err != nil {
		return err
	}
	modifiedObj, err := t.patch.Apply(rawObj)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(modifiedObj, &obj.Unstructured.Object)
	if err != nil {
		return err
	}
	return nil
}
