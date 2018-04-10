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
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kustomize/hash"
	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/kustomize/types"
)

// nameHashTransformer contains the prefix and the path config for each field that
// the name prefix will be applied.
type nameHashTransformer struct{}

var _ Transformer = &nameHashTransformer{}

// NewNameHashTransformer construct a nameHashTransformer.
func NewNameHashTransformer() Transformer {
	return &nameHashTransformer{}
}

// Transform appends hash to configmaps and secrets.
func (o *nameHashTransformer) Transform(m resource.ResourceCollection) error {
	for gvkn, obj := range m {
		switch {
		case types.SelectByGVK(gvkn.GVK, &schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}):
			appendHashForConfigMap(obj.Data)

		case types.SelectByGVK(gvkn.GVK, &schema.GroupVersionKind{Version: "v1", Kind: "Secret"}):
			appendHashForSecret(obj.Data)
		}
	}
	return nil
}

func appendHashForConfigMap(obj *unstructured.Unstructured) error {
	cm, err := unstructuredToConfigmap(obj)
	if err != nil {
		return err
	}
	h, err := hash.ConfigMapHash(cm)
	if err != nil {
		return err
	}
	nameWithHash := fmt.Sprintf("%s-%s", obj.GetName(), h)
	obj.SetName(nameWithHash)
	return nil
}

// TODO: Remove this function after we support hash unstructured objects
func unstructuredToConfigmap(in *unstructured.Unstructured) (*v1.ConfigMap, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out v1.ConfigMap
	err = json.Unmarshal(marshaled, &out)
	return &out, err
}

func appendHashForSecret(obj *unstructured.Unstructured) error {
	secret, err := unstructuredToSecret(obj)
	if err != nil {
		return err
	}
	h, err := hash.SecretHash(secret)
	if err != nil {
		return err
	}
	nameWithHash := fmt.Sprintf("%s-%s", obj.GetName(), h)
	obj.SetName(nameWithHash)
	return nil
}

// TODO: Remove this function after we support hash unstructured objects
func unstructuredToSecret(in *unstructured.Unstructured) (*v1.Secret, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out v1.Secret
	err = json.Unmarshal(marshaled, &out)
	return &out, err
}
