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

// Package resource implements representations of k8s API resources as "unstructured" objects.
package resource

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource is a Kubernetes Resource Object paired with a behavior.
type Resource struct {
	unstruct *unstructured.Unstructured
	behavior string
}

// NewResource returns a new instance of Resource.
func NewResource(u *unstructured.Unstructured, b string) *Resource {
	return &Resource{unstruct: u, behavior: b}
}

// NewBehaviorlessResource returns a new instance of Resource.
func NewBehaviorlessResource(u *unstructured.Unstructured) *Resource {
	return &Resource{unstruct: u}
}

// Behavior returns the behavior for the resource.
func (r *Resource) Behavior() string {
	return r.behavior
}

// Unstruct returns the unstructured object holding the resource.
func (r *Resource) Unstruct() *unstructured.Unstructured {
	return r.unstruct
}

// SetUnstruct sets a new member.
func (r *Resource) SetUnstruct(u *unstructured.Unstructured) {
	r.unstruct = u
}

// Id returns the ResId for the resource.
func (r *Resource) Id() ResId {
	var empty ResId
	if r.unstruct == nil {
		return empty
	}
	gvk := r.unstruct.GroupVersionKind()
	return NewResId(gvk, r.unstruct.GetName())
}

func (r *Resource) Merge(other *Resource) {
	r.Replace(other)
	mergeConfigmap(r.unstruct.Object, other.unstruct.Object, r.unstruct.Object)
}

func (r *Resource) Replace(other *Resource) {
	r.unstruct.SetLabels(mergeStringMaps(other.unstruct.GetLabels(), r.unstruct.GetLabels()))
	r.unstruct.SetAnnotations(mergeStringMaps(other.unstruct.GetAnnotations(), r.unstruct.GetAnnotations()))
	r.unstruct.SetName(other.unstruct.GetName())
}

// TODO: Add BinaryData once we sync to new k8s.io/api
func mergeConfigmap(mergedTo map[string]interface{}, maps ...map[string]interface{}) {
	mergedMap := map[string]interface{}{}
	for _, m := range maps {
		datamap, ok := m["data"].(map[string]interface{})
		if ok {
			for key, value := range datamap {
				mergedMap[key] = value
			}
		}
	}
	mergedTo["data"] = mergedMap
}

func mergeStringMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, m := range maps {
		for key, value := range m {
			result[key] = value
		}
	}
	return result
}

func newUnstructuredFromObject(in runtime.Object) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out unstructured.Unstructured
	err = out.UnmarshalJSON(marshaled)
	return &out, err
}
