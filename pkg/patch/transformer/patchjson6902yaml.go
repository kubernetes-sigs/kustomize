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
	"github.com/ghodss/yaml"
	"github.com/krishicks/yaml-patch"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
)

// patchJson6902YAMLTransformer applies patches.
type patchJson6902YAMLTransformer struct {
	target resource.ResId
	patch  yamlpatch.Patch
}

var _ transformers.Transformer = &patchJson6902YAMLTransformer{}

// newPatchJson6902YAMLTransformer constructs a PatchJson6902 transformer.
func newPatchJson6902YAMLTransformer(t resource.ResId, p yamlpatch.Patch) (transformers.Transformer, error) {
	if len(p) == 0 {
		return transformers.NewNoOpTransformer(), nil
	}
	return &patchJson6902YAMLTransformer{target: t, patch: p}, nil
}

// Transform apply the json patches on top of the base resources.
func (t *patchJson6902YAMLTransformer) Transform(baseResourceMap resmap.ResMap) error {
	obj, err := findTargetObj(baseResourceMap, t.target)
	if obj == nil {
		return err
	}
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
