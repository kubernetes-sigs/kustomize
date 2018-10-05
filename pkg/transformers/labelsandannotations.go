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
	"errors"
	"fmt"

	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformerconfig"
)

// mapTransformer contains a map string->string and path configs
// The map will be applied to the fields specified in path configs.
type mapTransformer struct {
	m           map[string]string
	pathConfigs []transformerconfig.PathConfig
}

var _ Transformer = &mapTransformer{}

// NewLabelsMapTransformer construct a mapTransformer with a given pathConfig slice
func NewLabelsMapTransformer(m map[string]string, p []transformerconfig.PathConfig) (Transformer, error) {
	return NewMapTransformer(p, m)
}

// NewAnnotationsMapTransformer construct a mapTransformer with a given pathConfig slice
func NewAnnotationsMapTransformer(m map[string]string, p []transformerconfig.PathConfig) (Transformer, error) {
	return NewMapTransformer(p, m)
}

// NewMapTransformer construct a mapTransformer.
func NewMapTransformer(pc []transformerconfig.PathConfig, m map[string]string) (Transformer, error) {
	if m == nil {
		return NewNoOpTransformer(), nil
	}
	if pc == nil {
		return nil, errors.New("pathConfigs is not expected to be nil")
	}
	return &mapTransformer{pathConfigs: pc, m: m}, nil
}

// Transform apply each <key, value> pair in the mapTransformer to the
// fields specified in mapTransformer.
func (o *mapTransformer) Transform(m resmap.ResMap) error {
	for id := range m {
		objMap := m[id].FunStruct().Map()
		for _, path := range o.pathConfigs {
			if !id.Gvk().IsSelected(&path.Gvk) {
				continue
			}
			err := mutateField(objMap, path.PathSlice(), path.CreateIfNotPresent, o.addMap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *mapTransformer) addMap(in interface{}) (interface{}, error) {
	m, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%#v is expectd to be %T", in, m)
	}
	for k, v := range o.m {
		m[k] = v
	}
	return m, nil
}
