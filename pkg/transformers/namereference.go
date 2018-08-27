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
	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// nameReferenceTransformer contains the referencing info between 2 GroupVersionKinds
type nameReferenceTransformer struct {
	pathConfigs []ReferencePathConfig
}

var _ Transformer = &nameReferenceTransformer{}

// NewDefaultingNameReferenceTransformer constructs a nameReferenceTransformer
// with defaultNameReferencepathConfigs.
func NewDefaultingNameReferenceTransformer() (Transformer, error) {
	return NewNameReferenceTransformer(defaultNameReferencePathConfigs)
}

// NewNameReferenceTransformer construct a nameReferenceTransformer.
func NewNameReferenceTransformer(pc []ReferencePathConfig) (Transformer, error) {
	if pc == nil {
		return nil, errors.New("pathConfigs is not expected to be nil")
	}
	return &nameReferenceTransformer{pathConfigs: pc}, nil
}

// Transform does the fields update according to pathConfigs.
// The old name is in the key in the map and the new name is in the object
// associated with the key. e.g. if <k, v> is one of the key-value pair in the map,
// then the old name is k.Name and the new name is v.GetName()
func (o *nameReferenceTransformer) Transform(m resmap.ResMap) error {
	for id := range m {
		objMap := m[id].UnstructuredContent()
		for _, referencePathConfig := range o.pathConfigs {
			for _, path := range referencePathConfig.pathConfigs {
				if !selectByGVK(id.Gvk(), path.GroupVersionKind) {
					continue
				}
				err := mutateField(objMap, path.Path, path.CreateIfNotPresent,
					o.updateNameReference(referencePathConfig.referencedGVK, m, id, path.Path))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (o *nameReferenceTransformer) updateNameReference(
	GVK schema.GroupVersionKind, m resmap.ResMap, referencingId resource.ResId, referencingPath []string) func(in interface{}) (interface{}, error) {
	return func(in interface{}) (interface{}, error) {
		s, ok := in.(string)
		if !ok {
			return nil, fmt.Errorf("%#v is expectd to be %T", in, s)
		}

		matchedIds := m.FindByGVKN(resource.NewResId(GVK, s))
		if len(matchedIds) == 0 {
			return in, nil
		}
		if len(matchedIds) > 1 {
			return nil, fmt.Errorf("found multiple objects %#v matching name reference %#v path %s",
				matchedIds, referencingId, strings.Join(referencingPath, "."))
		}

		res, _ := m.DemandOneMatchForId(matchedIds[0])
		return res.GetName(), nil
	}
}
