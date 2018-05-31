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
	"bytes"
	"encoding/json"
	"sort"

	"github.com/ghodss/yaml"

	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

func objectToUnstructured(in runtime.Object) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out unstructured.Unstructured
	err = out.UnmarshalJSON(marshaled)
	return &out, err
}

// ResourceCollection is a map from GroupVersionKindName to Resource
type ResourceCollection map[types.GroupVersionKindName]*Resource

// EncodeAsYaml encodes the map `in` and output the encoded objects separated by `---`.
func (in ResourceCollection) EncodeAsYaml() ([]byte, error) {
	gvknList := []types.GroupVersionKindName{}
	for gvkn := range in {
		gvknList = append(gvknList, gvkn)
	}
	sort.Sort(types.ByGVKN(gvknList))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, gvkn := range gvknList {
		obj := in[gvkn].Data
		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if !firstObj {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.Write(out)
		if err != nil {
			return nil, err
		}
		firstObj = false
	}
	return buf.Bytes(), nil
}
