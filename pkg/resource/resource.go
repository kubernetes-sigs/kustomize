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
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
	internal "sigs.k8s.io/kustomize/pkg/internal/error"
	"sigs.k8s.io/kustomize/pkg/patch"
	"sigs.k8s.io/kustomize/pkg/resid"
)

// Resource is an "Unstructured" (json/map form) Kubernetes API resource object
// paired with a GenerationBehavior.
type Resource struct {
	unstructured.Unstructured
	b ifc.GenerationBehavior
}

// NewResourceWithBehavior returns a new instance of Resource.
func NewResourceWithBehavior(obj runtime.Object, b ifc.GenerationBehavior) (*Resource, error) {
	// Convert obj to a byte stream, then convert that to JSON (Unstructured).
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var u unstructured.Unstructured
	err = u.UnmarshalJSON(marshaled)
	// creationTimestamp always 'null', remove it
	u.SetCreationTimestamp(metav1.Time{})
	return &Resource{Unstructured: u, b: b}, err
}

// NewResourceFromMap returns a new instance of Resource.
func NewResourceFromMap(m map[string]interface{}) *Resource {
	return NewResourceFromUnstruct(unstructured.Unstructured{Object: m})
}

// NewResourceFromUnstruct returns a new instance of Resource.
func NewResourceFromUnstruct(u unstructured.Unstructured) *Resource {
	return &Resource{Unstructured: u, b: ifc.BehaviorUnspecified}
}

// NewResourceSliceFromPatches returns a slice of resources given a patch path
// slice from a kustomization file.
func NewResourceSliceFromPatches(
	ldr ifc.Loader, paths []patch.StrategicMerge,
	decoder ifc.Decoder) ([]*Resource, error) {
	var result []*Resource
	for _, path := range paths {
		content, err := ldr.Load(string(path))
		if err != nil {
			return nil, err
		}
		res, err := NewResourceSliceFromBytes(content, decoder)
		if err != nil {
			return nil, internal.Handler(err, string(path))
		}
		result = append(result, res...)
	}
	return result, nil
}

// NewResourceSliceFromBytes unmarshalls bytes into a Resource slice.
func NewResourceSliceFromBytes(
	in []byte, decoder ifc.Decoder) ([]*Resource, error) {
	decoder.SetInput(in)
	var result []*Resource
	var err error
	for err == nil || isEmptyYamlError(err) {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err == nil {
			result = append(result, NewResourceFromUnstruct(out))
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
}

// String returns resource as JSON.
func (r *Resource) String() string {
	bs, err := r.MarshalJSON()
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return r.b.String() + ":" + strings.TrimSpace(string(bs))
}

// Behavior returns the behavior for the resource.
func (r *Resource) Behavior() ifc.GenerationBehavior {
	return r.b
}

// SetBehavior changes the resource to the new behavior
func (r *Resource) SetBehavior(b ifc.GenerationBehavior) *Resource {
	r.b = b
	return r
}

// IsGenerated checks if the resource is generated from a generator
func (r *Resource) IsGenerated() bool {
	return r.b != ifc.BehaviorUnspecified
}

// Id returns the ResId for the resource.
func (r *Resource) Id() resid.ResId {
	return resid.NewResId(gvk.FromSchemaGvk(r.GroupVersionKind()), r.GetName())
}

// Merge performs merge with other resource.
func (r *Resource) Merge(other *Resource) {
	r.Replace(other)
	mergeConfigmap(r.Object, other.Object, r.Object)
}

// Replace performs replace with other resource.
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

// GetFieldValue returns value at the given fieldpath.
func (r *Resource) GetFieldValue(fieldPath string) (string, error) {
	return getFieldValue(r.UnstructuredContent(), strings.Split(fieldPath, "."))
}

func getFieldValue(m map[string]interface{}, pathToField []string) (string, error) {
	if len(pathToField) == 0 {
		return "", fmt.Errorf("field not found")
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
	v := m[pathToField[0]]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return getFieldValue(typedV, pathToField[1:])
	default:
		return "", fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}
