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

	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/kustomize/types"
)

// namePrefixTransformer contains the prefix and the path config for each field that
// the name prefix will be applied.
type namePrefixTransformer struct {
	prefix      string
	pathConfigs []PathConfig
}

var _ Transformer = &namePrefixTransformer{}

var defaultNamePrefixPathConfigs = []PathConfig{
	{
		Path:               []string{"metadata", "name"},
		CreateIfNotPresent: false,
	},
}

// NewDefaultingNamePrefixTransformer construct a namePrefixTransformer with defaultNamePrefixPathConfigs.
func NewDefaultingNamePrefixTransformer(nameprefix string) (Transformer, error) {
	return NewNamePrefixTransformer(defaultNamePrefixPathConfigs, nameprefix)
}

// NewNamePrefixTransformer construct a namePrefixTransformer.
func NewNamePrefixTransformer(pc []PathConfig, np string) (Transformer, error) {
	if len(np) == 0 {
		return NewNoOpTransformer(), nil
	}
	if pc == nil {
		return nil, errors.New("pathConfigs is not expected to be nil")
	}
	return &namePrefixTransformer{pathConfigs: pc, prefix: np}, nil
}

// Transform prepends the name prefix.
func (o *namePrefixTransformer) Transform(m resource.ResourceCollection) error {
	for gvkn := range m {
		obj := m[gvkn].Data
		objMap := obj.UnstructuredContent()
		for _, path := range o.pathConfigs {
			if !types.SelectByGVK(gvkn.GVK, path.GroupVersionKind) {
				continue
			}
			err := mutateField(objMap, path.Path, path.CreateIfNotPresent, o.addPrefix)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *namePrefixTransformer) addPrefix(in interface{}) (interface{}, error) {
	s, ok := in.(string)
	if !ok {
		return nil, fmt.Errorf("%#v is expectd to be %T", in, s)
	}
	return o.prefix + s, nil
}
