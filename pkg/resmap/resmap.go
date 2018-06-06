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

// Package resmap implements a map from ResId to Resource that tracks all resources in a kustomization.
package resmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// ResMap is a map from ResId to Resource.
type ResMap map[resource.ResId]*resource.Resource

// EncodeAsYaml encodes a ResMap to YAML; encoded objects separated by `---`.
func (m ResMap) EncodeAsYaml() ([]byte, error) {
	ids := []resource.ResId{}
	for gvkn := range m {
		ids = append(ids, gvkn)
	}
	sort.Sort(IdSlice(ids))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, id := range ids {
		obj := m[id].Unstruct()
		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if firstObj {
			firstObj = false
		} else {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.Write(out)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (m1 ResMap) ErrorIfNotEqual(m2 ResMap) error {
	if len(m1) != len(m2) {
		keySet1 := []resource.ResId{}
		keySet2 := []resource.ResId{}
		for id := range m1 {
			keySet1 = append(keySet1, id)
		}
		for id := range m2 {
			keySet2 = append(keySet2, id)
		}
		return fmt.Errorf("maps has different number of entries: %#v doesn't equals %#v", keySet1, keySet2)
	}
	for id, obj1 := range m1 {
		obj2, found := m2[id]
		if !found {
			return fmt.Errorf("%#v doesn't exist in %#v", id, m2)
		}
		if !reflect.DeepEqual(obj1.Unstruct(), obj2.Unstruct()) {
			return fmt.Errorf("%#v doesn't match %#v", obj1.Unstruct(), obj2.Unstruct())
		}
	}
	return nil
}

func (m ResMap) insert(newName string, obj *unstructured.Unstructured) error {
	oldName := obj.GetName()
	gvk := obj.GroupVersionKind()
	id := resource.NewResId(gvk, oldName)

	if _, found := m[id]; found {
		return fmt.Errorf("The <name: %q, GroupVersionKind: %v> already exists in the map", oldName, gvk)
	}
	obj.SetName(newName)
	m[id] = resource.NewBehaviorlessResource(obj)
	return nil
}

// NewResourceSliceFromPatches returns a slice of Resources given a patch path slice from kustomization file.
func NewResourceSliceFromPatches(
	loader loader.Loader, paths []string) ([]*resource.Resource, error) {
	result := []*resource.Resource{}
	for _, path := range paths {
		content, err := loader.Load(path)
		if err != nil {
			return nil, err
		}

		res, err := newResourceSliceFromBytes(content)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}
	return result, nil
}

// NewResMapFromFiles returns a ResMap given a resource path slice.
func NewResMapFromFiles(loader loader.Loader, paths []string) (ResMap, error) {
	result := []ResMap{}
	for _, path := range paths {
		content, err := loader.Load(path)
		if err != nil {
			return nil, err
		}

		res, err := newResMapFromBytes(content)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return Merge(result...)
}

// newResMapFromBytes decodes a list of objects in byte array format.
func newResMapFromBytes(b []byte) (ResMap, error) {
	resources, err := newResourceSliceFromBytes(b)
	if err != nil {
		return nil, err
	}

	result := ResMap{}
	for _, res := range resources {
		gvkn := res.Id()
		if _, found := result[gvkn]; found {
			return result, fmt.Errorf("GroupVersionKindName: %#v already exists b the map", gvkn)
		}
		result[gvkn] = res
	}
	return result, nil
}

func newResMapFromResourceSlice(resources []*resource.Resource) (ResMap, error) {
	result := ResMap{}
	for _, res := range resources {
		gvkn := res.Id()
		if _, found := result[gvkn]; found {
			return nil, fmt.Errorf("duplicated %#v is not allowed", gvkn)
		}
		result[gvkn] = res
	}
	return result, nil
}

func newResourceSliceFromBytes(in []byte) ([]*resource.Resource, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	result := []*resource.Resource{}

	var err error
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			break
		}
		result = append(result, resource.NewBehaviorlessResource(&out))
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

// Merge combines many maps to one.
func Merge(maps ...ResMap) (ResMap, error) {
	result := ResMap{}
	for _, m := range maps {
		for gvkn, obj := range m {
			if _, found := result[gvkn]; found {
				return nil, fmt.Errorf("there is already an entry: %q", gvkn)
			}
			result[gvkn] = obj
		}
	}

	return result, nil
}

const behaviorCreate = "create"
const behaviorReplace = "replace"
const behaviorMerge = "merge"

// MergeWithOverride merges the entries in the ResMap slice with Override.
// If there is already an entry with the same Id , different actions are performed
// according to value of behavior field:
// 'create': create a new one;
// 'replace': replace the data only; keep the labels and annotations
// 'merge': merge the data; keep the labels and annotations
func MergeWithOverride(maps ...ResMap) (ResMap, error) {
	result := ResMap{}
	for _, m := range maps {
		for gvkn, resource := range m {
			if _, found := result[gvkn]; found {
				switch resource.Behavior() {
				case "", behaviorCreate:
					return nil, fmt.Errorf("Create an existing gvkn %#v is not allowed", gvkn)
				case behaviorReplace:
					glog.V(4).Infof("Replace object %v by %v", result[gvkn].Unstruct().Object, resource.Unstruct().Object)
					resource.Replace(result[gvkn])
					result[gvkn] = resource
				case behaviorMerge:
					glog.V(4).Infof("Merge object %v with %v", result[gvkn].Unstruct().Object, resource.Unstruct().Object)
					resource.Merge(result[gvkn])
					result[gvkn] = resource
					glog.V(4).Infof("The merged object is %v", result[gvkn].Unstruct().Object)
				default:
					return nil, fmt.Errorf("The behavior of %#v must be one of merge and replace since it already exists in the base", gvkn)
				}
			} else {
				switch resource.Behavior() {
				case "", behaviorCreate:
					result[gvkn] = resource
				case behaviorMerge, behaviorReplace:
					return nil, fmt.Errorf("No merge or replace is allowed for non existing gvkn %#v", gvkn)
				default:
					return nil, fmt.Errorf("The behavior of %#v must be create since it doesn't exist", gvkn)
				}
			}
		}
	}
	return result, nil
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
