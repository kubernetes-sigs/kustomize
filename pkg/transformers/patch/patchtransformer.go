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

package patch

import (
	"fmt"
	"reflect"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
)

// patchTransformer applies different types of patches.
type patchTransformer struct {
	transformers []transformers.Transformer
}

var _ transformers.Transformer = &patchTransformer{}

// NewPatchTransformer constructs a patchTransformer.
func NewPatchTransformer(slice []*resource.Resource, patchesJ6 map[resource.ResId][]byte) (transformers.Transformer, error) {
	var ts []transformers.Transformer
	patchSMt, err := NewPatchStrategicMergeTransformer(slice)
	if err != nil {
		return nil, err
	}
	ts = append(ts, patchSMt)
	patchJ6t, err := NewPatchJson6902Transformer(patchesJ6)
	if err != nil {
		return nil, err
	}
	ts = append(ts, patchJ6t)
	return &patchTransformer{ts}, nil
}

// Transform apply the patches on top of the base resources.
func (pt *patchTransformer) Transform(baseResourceMap resmap.ResMap) error {
	return pt.transformWithCheckConflicts(baseResourceMap)
}

func (pt *patchTransformer) transformWithCheckConflicts(m resmap.ResMap) error {
	mcopy := resmap.ResMap{}
	for id, res := range m {
		mcopy[id] = &resource.Resource{Unstructured: res.Unstructured}
		mcopy[id].SetBehavior(res.Behavior())
	}
	mt1 := transformers.NewMultiTransformer(pt.transformers)
	mt2 := transformers.NewMultiTransformer(pt.reverse())
	err := mt1.Transform(m)
	if err != nil {
		return err
	}
	err = mt2.Transform(mcopy)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(m, mcopy) {
		return fmt.Errorf("There is conflict between different types of patches.")
	}
	return nil
}

func (pt *patchTransformer) reverse() []transformers.Transformer {
	var result []transformers.Transformer
	for _, t := range pt.transformers {
		result = append([]transformers.Transformer{t}, result...)
	}
	return result
}
