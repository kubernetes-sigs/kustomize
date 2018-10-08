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
	"log"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/internal/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/ifc"
	internal "sigs.k8s.io/kustomize/pkg/internal/error"
	"sigs.k8s.io/kustomize/pkg/patch"
	"sigs.k8s.io/kustomize/pkg/resid"
)

// Resource is map representation of a Kubernetes API resource object
// paired with a GenerationBehavior.
type Resource struct {
	ifc.Kunstructured
	b ifc.GenerationBehavior
}

// NewWithBehavior returns a new instance of Resource.
func NewWithBehavior(obj runtime.Object, b ifc.GenerationBehavior) (*Resource, error) {
	u, err := k8sdeps.NewKunstructuredFromObject(obj)
	return &Resource{Kunstructured: u, b: b}, err
}

// NewFromMap returns a new instance of Resource.
func NewFromMap(m map[string]interface{}) *Resource {
	return &Resource{
		Kunstructured: k8sdeps.NewKunstructuredFromMap(m),
		b:             ifc.BehaviorUnspecified}
}

// NewFromKunstructured returns a new instance of Resource.
func NewFromKunstructured(u ifc.Kunstructured) *Resource {
	if u == nil {
		log.Fatal("unstruct ifc must not be null")
	}
	return &Resource{Kunstructured: u, b: ifc.BehaviorUnspecified}
}

// NewSliceFromPatches returns a slice of resources given a patch path
// slice from a kustomization file.
func NewSliceFromPatches(
	ldr ifc.Loader, paths []patch.StrategicMerge,
	decoder ifc.Decoder) ([]*Resource, error) {
	var result []*Resource
	for _, path := range paths {
		content, err := ldr.Load(string(path))
		if err != nil {
			return nil, err
		}
		res, err := NewSliceFromBytes(content, decoder)
		if err != nil {
			return nil, internal.Handler(err, string(path))
		}
		result = append(result, res...)
	}
	return result, nil
}

// NewSliceFromBytes unmarshalls bytes into a Resource slice.
func NewSliceFromBytes(
	in []byte, decoder ifc.Decoder) ([]*Resource, error) {
	funStructs, err := k8sdeps.NewKunstructuredSliceFromBytes(in, decoder)
	if err != nil {
		return nil, err
	}
	var result []*Resource
	for _, u := range funStructs {
		result = append(result, NewFromKunstructured(u))
	}
	return result, nil
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
	return resid.NewResId(r.GetGvk(), r.GetName())
}

// Merge performs merge with other resource.
func (r *Resource) Merge(other *Resource) {
	r.Replace(other)
	mergeConfigmap(r.Map(), other.Map(), r.Map())
}

// Replace performs replace with other resource.
func (r *Resource) Replace(other *Resource) {
	r.SetLabels(mergeStringMaps(other.GetLabels(), r.GetLabels()))
	r.SetAnnotations(
		mergeStringMaps(other.GetAnnotations(), r.GetAnnotations()))
	r.SetName(other.GetName())
}

// TODO: Add BinaryData once we sync to new k8s.io/api
func mergeConfigmap(
	mergedTo map[string]interface{},
	maps ...map[string]interface{}) {
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
