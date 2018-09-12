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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/patch"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

// PatchJson6902Factory makes PatchJson6902 transformers
type PatchJson6902Factory struct {
	loader loader.Loader
}

// NewPatchJson6902Factory returns a new PatchJson6902Factory.
func NewPatchJson6902Factory(l loader.Loader) PatchJson6902Factory {
	return PatchJson6902Factory{loader: l}
}

// MakePatchJson6902Transformer returns a transformer for applying Json6902 patch
func (f PatchJson6902Factory) MakePatchJson6902Transformer(patches []patch.PatchJson6902) (transformers.Transformer, error) {
	var ts []transformers.Transformer
	for _, p := range patches {
		t, err := f.makeOnePatchJson6902Transformer(p)
		if err != nil {
			return nil, err
		}
		if t != nil {
			ts = append(ts, t)
		}
	}
	return transformers.NewMultiTransformerWithConflictCheck(ts), nil
}

func (f PatchJson6902Factory) makeOnePatchJson6902Transformer(p patch.PatchJson6902) (transformers.Transformer, error) {
	if p.Target == nil {
		return nil, fmt.Errorf("must specify the target field in patchesJson6902")
	}
	if p.Path == "" {
		return nil, fmt.Errorf("must specify the path for a json patch file")
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

	rawOp, err := f.loader.Load(p.Path)
	if err != nil {
		return nil, err
	}
	if isJsonFormat(rawOp) {
		decodedPatch, err := jsonpatch.DecodePatch(rawOp)
		if err != nil {
			return nil, err
		}
		return newPatchJson6902JSONTransformer(targetId, decodedPatch)
	}
	decodedPatch, err := yamlpatch.DecodePatch(rawOp)
	if err != nil {
		return nil, err
	}
	return newPatchJson6902YAMLTransformer(targetId, decodedPatch)
}

func isJsonFormat(data []byte) bool {
	return data[0] == '['
}
