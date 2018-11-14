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

// nameSuffixTransformer contains the suffix and the FieldSpecs
// for each field needing a name suffix.
type nameSuffixTransformer struct {
	suffix           string
	fieldSpecsToUse  []config.FieldSpec
	fieldSpecsToSkip []config.FieldSpec
}

var _ Transformer = &nameSuffixTransformer{}

var suffixFieldSpecsToSkip = []config.FieldSpec{
	{
		Gvk: gvk.Gvk{Kind: "CustomResourceDefinition"},
	},
}

// deprecateNameSuffixFieldSpec will be moved into suffixFieldSpecsToSkip in next release
var deprecateNameSuffixFieldSpec = config.FieldSpec{
	Gvk: gvk.Gvk{Kind: "Namespace"},
}

// NewNameSuffixTransformer construct a nameSuffixTransformer.
func NewNameSuffixTransformer(ns string, pc []config.FieldSpec) (Transformer, error) {
	if len(ns) == 0 {
		return NewNoOpTransformer(), nil
	}
	if pc == nil {
		return nil, errors.New("fieldSpecs is not expected to be nil")
	}
	return &nameSuffixTransformer{fieldSpecsToUse: pc, suffix: ns, fieldSpecsToSkip: suffixFieldSpecsToSkip}, nil
}

// Transform appends the name suffix.
func (o *nameSuffixTransformer) Transform(m resmap.ResMap) error {
	// Fill map "mf" with entries subject to name modification, and
	// delete these entries from "m", so that for now m retains only
	// the entries whose names will not be modified.
	mf := resmap.ResMap{}
	for id := range m {
		found := false
		for _, path := range o.fieldSpecsToSkip {
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
		if id.Gvk().IsSelected(&deprecateNameSuffixFieldSpec.Gvk) {
			log.Println("Adding name suffix to Namespace resource will be deprecated in next release.")
		}
		objMap := mf[id].Map()
		for _, path := range o.fieldSpecsToUse {
			if !id.Gvk().IsSelected(&path.Gvk) {
				continue
			}
			err := mutateField(objMap, path.PathSlice(), path.CreateIfNotPresent, o.addSuffix)
			if err != nil {
				return err
			}
			newId := id.CopyWithNewSuffix(o.suffix)
			m[newId] = mf[id]
		}
	}
	return nil
}

func (o *nameSuffixTransformer) addSuffix(in interface{}) (interface{}, error) {
	s, ok := in.(string)
	if !ok {
		return nil, fmt.Errorf("%#v is expectd to be %T", in, s)
	}
	return s + o.suffix, nil
}
