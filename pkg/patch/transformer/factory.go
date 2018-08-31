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

	"github.com/evanphx/json-patch"
	"github.com/krishicks/yaml-patch"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/patch"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PatchJson6902Factory makes PatchJson6902 transformers.
type PatchJson6902Factory struct {
	targetId       resource.ResId
	operationsYAML yamlpatch.Patch
	operationsJSON jsonpatch.Patch
}

// NewPatchJson6902Factory returns a new PatchJson6902Factory.
func NewPatchJson6902Factory(l loader.Loader, p patch.PatchJson6902) (*PatchJson6902Factory, error) {

	if p.Target == nil {
		return nil, fmt.Errorf("must specify the target field in patchesJson6902")
	}
	if p.Path != "" && p.JsonPatch != nil {
		return nil, fmt.Errorf("cannot specify path and jsonPath at the same time")
	}

	targetId := resource.NewResIdWithPrefixNamespace(
		schema.GroupVersionKind{
			Group:   p.Target.Group,
			Version: p.Target.Version,
			Kind:    p.Target.Kind,
		},
		p.Target.Name,
		"",
		p.Target.Namespace,
	)

	if p.JsonPatch != nil {
		return &PatchJson6902Factory{targetId: targetId, operationsYAML: p.JsonPatch}, nil
	}
	if p.Path != "" {
		rawOp, err := l.Load(p.Path)
		if err != nil {
			return nil, err
		}
		patch, err := jsonpatch.DecodePatch(rawOp)
		if err != nil {
			return nil, err
		}
		return &PatchJson6902Factory{targetId: targetId, operationsJSON: patch}, nil
	}
	return nil, nil
}

// MakePatchJson6902Transformer returns a transformer for applying Json6902 patch
func (f *PatchJson6902Factory) MakePatchJson6902Transformer() (transformers.Transformer, error) {
	if f.operationsJSON != nil {
		return newPatchJson6902JSONTransformer(f.targetId, f.operationsJSON)
	}
	return newPatchJson6902YAMLTransformer(f.targetId, f.operationsYAML)
}
