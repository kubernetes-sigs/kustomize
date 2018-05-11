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

package resource

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/kustomize/types"
)

// Resource represents a Kubernetes Resource Object for ex. Deployment, Server
// ConfigMap etc.
type Resource struct {
	Data     *unstructured.Unstructured
	Behavior string
}

// GVKN returns Group/Version/Kind/Name for the resource.
func (r *Resource) GVKN() types.GroupVersionKindName {
	var emptyZVKN types.GroupVersionKindName
	if r.Data == nil {
		return emptyZVKN
	}
	gvk := r.Data.GroupVersionKind()
	return types.GroupVersionKindName{GVK: gvk, Name: r.Data.GetName()}
}

// ResourceCollection is a map from GroupVersionKindName to Resource
type ResourceCollection map[types.GroupVersionKindName]*Resource

func objectToUnstructured(in runtime.Object) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out unstructured.Unstructured
	err = out.UnmarshalJSON(marshaled)
	return &out, err
}
