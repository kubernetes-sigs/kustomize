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

	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/patch"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
)

// NewPatchJson6902Transformer constructs a PatchJson6902 transformer.
func NewPatchJson6902Transformer(l loader.Loader, p patch.PatchJson6902) (transformers.Transformer, error) {
	if p.Target == nil {
		return nil, fmt.Errorf("must specify the target field in patchesJson6902")
	}
	if p.Path != "" && p.JsonPatch != nil {
		return nil, fmt.Errorf("cannot specify path and jsonPath at the same time")
	}

	if p.JsonPatch != nil {
		return NewPatchJson6902YAMLTransformer(p.Target, p.JsonPatch)
	}
	if p.Path != "" {
		operations, err := l.Load(p.Path)
		if err != nil {
			return nil, err
		}
		return NewPatchJson6902JSONTransformer(p.Target, operations)
	}

	return transformers.NewNoOpTransformer(), nil
}
