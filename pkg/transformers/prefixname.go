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

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// namePrefixTransformer contains the prefix and the path config for each field that
// the name prefix will be applied.
type namePrefixTransformer struct {
	prefix          string
	pathConfigs     []PathConfig
	skipPathConfigs []PathConfig
}

var _ Transformer = &namePrefixTransformer{}

var defaultNamePrefixPathConfigs = []PathConfig{
	{
		Path:               []string{"metadata", "name"},
		CreateIfNotPresent: false,
	},
}

var skipNamePrefixPathConfigs = []PathConfig{
	{
		GroupVersionKind: &schema.GroupVersionKind{Kind: "CustomResourceDefinition"},
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
	return &namePrefixTransformer{pathConfigs: pc, prefix: np, skipPathConfigs: skipNamePrefixPathConfigs}, nil
}

// Transform prepends the name prefix.
func (o *namePrefixTransformer) Transform(m resmap.ResMap) error {
	mf := resmap.ResMap{}

	for id := range m {
		mf[id] = m[id]
		for _, path := range o.skipPathConfigs {
			if selectByGVK(id.Gvk(), path.GroupVersionKind) {
				delete(mf, id)
				break
			}
		}
	}

	for id := range mf {
		objMap := mf[id].UnstructuredContent()
		for _, path := range o.pathConfigs {
			if !selectByGVK(id.Gvk(), path.GroupVersionKind) {
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

// AddPrefixPathConfigs adds extra path configs to the default one
func AddPrefixPathConfigs(pathConfigs ...PathConfig) {
	defaultNamePrefixPathConfigs = append(defaultNamePrefixPathConfigs, pathConfigs...)
}
