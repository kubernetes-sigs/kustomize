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
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource is an "Unstructured" (json/map form) Kubernetes API resource object
// paired with a GenerationBehavior.
type Resource struct {
	unstructured.Unstructured
	b GenerationBehavior
}

// NewResourceWithBehavior returns a new instance of Resource.
func NewResourceWithBehavior(obj runtime.Object, b GenerationBehavior) (*Resource, error) {
	// Convert obj to a byte stream, then convert that to JSON (Unstructured).
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var u unstructured.Unstructured
	err = u.UnmarshalJSON(marshaled)
	return &Resource{Unstructured: u, b: b}, nil
}

// NewBehaviorlessResource returns a new instance of Resource.
func NewBehaviorlessResource(u *unstructured.Unstructured) *Resource {
	return &Resource{Unstructured: *u, b: BehaviorUnspecified}
}

// Behavior returns the behavior for the resource.
func (r *Resource) Behavior() GenerationBehavior {
	return r.b
}

// Id returns the ResId for the resource.
func (r *Resource) Id() ResId {
	return NewResId(r.GroupVersionKind(), r.GetName())
}

func (r *Resource) Merge(other *Resource) {
	r.Replace(other)
	mergeConfigmap(r.Object, other.Object, r.Object)
}

func (r *Resource) Replace(other *Resource) {
	r.SetLabels(mergeStringMaps(other.GetLabels(), r.GetLabels()))
	r.SetAnnotations(mergeStringMaps(other.GetAnnotations(), r.GetAnnotations()))
	r.SetName(other.GetName())
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

func GetFieldValue(m map[string]interface{}, pathToField []string) (string, error) {
	if len(pathToField) == 0 {
		return "", fmt.Errorf("Field not found")
	}

	if len(pathToField) == 1 {
		if v, found := m[pathToField[0]]; found {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return "", fmt.Errorf("value at fieldpath is not of string type")
		}
		return "", fmt.Errorf("field at given fieldpath does not exist")
	}

	curr, rest := pathToField[0], pathToField[1]

	v := m[curr]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return GetFieldValue(typedV, []string{rest})
	default:
		return "", fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}
