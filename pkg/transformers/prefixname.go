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
	"log"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

// namePrefixTransformer contains the prefix and the path config for each field that
// the name prefix will be applied.
type namePrefixTransformer struct {
	prefix          string
	pathConfigs     []config.PathConfig
	skipPathConfigs []config.PathConfig
}

var _ Transformer = &namePrefixTransformer{}

var skipNamePrefixPathConfigs = []config.PathConfig{
	{
		Gvk: gvk.Gvk{Kind: "CustomResourceDefinition"},
	},
}

// deprecateNamePrefixPathConfig will be moved into skipNamePrefixPathConfigs in next release
var deprecateNamePrefixPathConfig = config.PathConfig{
	Gvk: gvk.Gvk{Kind: "Namespace"},
}

// NewNamePrefixTransformer construct a namePrefixTransformer.
func NewNamePrefixTransformer(np string, pc []config.PathConfig) (Transformer, error) {
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
		found := false
		for _, path := range o.skipPathConfigs {
			if id.Gvk().IsSelected(&path.Gvk) {
				found = true
				break
			}
		}
		if !found {
			mf[id] = m[id]
			delete(m, id)
		}
	}

	for id := range mf {
		if id.Gvk().IsSelected(&deprecateNamePrefixPathConfig.Gvk) {
			log.Println("Adding nameprefix to Namespace resource will be deprecated in next release.")
		}
		objMap := mf[id].Map()
		for _, path := range o.pathConfigs {
			if !id.Gvk().IsSelected(&path.Gvk) {
				continue
			}
			err := mutateField(objMap, path.PathSlice(), path.CreateIfNotPresent, o.addPrefix)
			if err != nil {
				return err
			}
			newId := id.CopyWithNewPrefix(o.prefix)
			m[newId] = mf[id]
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
