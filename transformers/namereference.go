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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/kustomize/types"
)

// nameReferenceTransformer contains the referencing info between 2 GroupVersionKinds
type nameReferenceTransformer struct {
	pathConfigs []referencePathConfig
}

var _ Transformer = &nameReferenceTransformer{}

// NewDefaultingNameReferenceTransformer constructs a nameReferenceTransformer
// with defaultNameReferencepathConfigs.
func NewDefaultingNameReferenceTransformer() (Transformer, error) {
	return NewNameReferenceTransformer(defaultNameReferencePathConfigs)
}

// NewNameReferenceTransformer construct a nameReferenceTransformer.
func NewNameReferenceTransformer(pc []referencePathConfig) (Transformer, error) {
	if pc == nil {
		return nil, errors.New("pathConfigs is not expected to be nil")
	}
	return &nameReferenceTransformer{pathConfigs: pc}, nil
}

// Transform does the fields update according to pathConfigs.
// The old name is in the key in the map and the new name is in the object
// associated with the key. e.g. if <k, v> is one of the key-value pair in the map,
// then the old name is k.Name and the new name is v.GetName()
func (o *nameReferenceTransformer) Transform(
	m resource.ResourceCollection) error {
	for GVKn := range m {
		obj := m[GVKn].Data
		objMap := obj.UnstructuredContent()
		for _, referencePathConfig := range o.pathConfigs {
			for _, path := range referencePathConfig.pathConfigs {
				if !types.SelectByGVK(GVKn.GVK, path.GroupVersionKind) {
					continue
				}
				err := mutateField(objMap, path.Path, path.CreateIfNotPresent,
					o.updateNameReference(referencePathConfig.referencedGVK, m))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// noMatchingGVKNError indicates failing to find a gvkn.GroupVersionKindName.
type noMatchingGVKNError struct {
	message string
}

// newNoMatchingGVKNError constructs an instance of noMatchingGVKNError with
// a given error message.
func newNoMatchingGVKNError(errMsg string) noMatchingGVKNError {
	return noMatchingGVKNError{errMsg}
}

// Error returns the error in string format.
func (err noMatchingGVKNError) Error() string {
	return err.message
}

func (o *nameReferenceTransformer) updateNameReference(
	GVK schema.GroupVersionKind,
	m resource.ResourceCollection,
) func(in interface{}) (interface{}, error) {
	return func(in interface{}) (interface{}, error) {
		s, ok := in.(string)
		if !ok {
			return nil, fmt.Errorf("%#v is expectd to be %T", in, s)
		}

		for GVKn, obj := range m {
			if !types.SelectByGVK(GVKn.GVK, &GVK) {
				continue
			}
			if GVKn.Name == s {
				return obj.Data.GetName(), nil
			}
		}
		return in, nil
	}
}
