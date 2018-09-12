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
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	internal "sigs.k8s.io/kustomize/pkg/internal/error"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/patch"
	"sigs.k8s.io/kustomize/pkg/resource"
)

// ResMap is a map from ResId to Resource.
type ResMap map[resource.ResId]*resource.Resource

// FindByGVKN find the matched ResIds by Group/Version/Kind and Name
func (m ResMap) FindByGVKN(inputId resource.ResId) []resource.ResId {
	var result []resource.ResId
	for id := range m {
		if id.GvknEquals(inputId) {
			result = append(result, id)
		}
	}
	return result
}

// DemandOneMatchForId find the matched resource by Group/Version/Kind and Name
func (m ResMap) DemandOneMatchForId(inputId resource.ResId) (*resource.Resource, bool) {
	result := m.FindByGVKN(inputId)
	if len(result) == 1 {
		return m[result[0]], true
	}
	return nil, false
}

// EncodeAsYaml encodes a ResMap to YAML; encoded objects separated by `---`.
func (m ResMap) EncodeAsYaml() ([]byte, error) {
	var ids []resource.ResId
	for id := range m {
		ids = append(ids, id)
	}
	sort.Sort(IdSlice(ids))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, id := range ids {
		obj := m[id]
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

// ErrorIfNotEqual returns error if maps are not equal.
func (m ResMap) ErrorIfNotEqual(m2 ResMap) error {
	if len(m) != len(m2) {
		var keySet1 []resource.ResId
		var keySet2 []resource.ResId
		for id := range m {
			keySet1 = append(keySet1, id)
		}
		for id := range m2 {
			keySet2 = append(keySet2, id)
		}
		return fmt.Errorf("maps has different number of entries: %#v doesn't equals %#v", keySet1, keySet2)
	}
	for id, obj1 := range m {
		obj2, found := m2[id]
		if !found {
			return fmt.Errorf("%#v doesn't exist in %#v", id, m2)
		}
		if !reflect.DeepEqual(obj1, obj2) {
			return fmt.Errorf("%#v doesn't deep equal %#v", obj1, obj2)
		}
	}
	return nil
}

// DeepCopy clone the resmap into a new one
func (m ResMap) DeepCopy() ResMap {
	mcopy := make(ResMap)
	for id, obj := range m {
		mcopy[id] = resource.NewResourceFromUnstruct(obj.Unstructured)
		mcopy[id].SetBehavior(obj.Behavior())
	}
	return mcopy
}

func (m ResMap) insert(newName string, obj *unstructured.Unstructured) error {
	oldName := obj.GetName()
	gvk := obj.GroupVersionKind()
	id := resource.NewResId(gvk, oldName)

	if _, found := m[id]; found {
		return fmt.Errorf("the <name: %q, GroupVersionKind: %v> already exists in the map", oldName, gvk)
	}
	obj.SetName(newName)
	m[id] = resource.NewResourceFromUnstruct(*obj)
	return nil
}

// FilterBy returns a ResMap containing ResIds with the same namespace and nameprefix
// with the inputId
func (m ResMap) FilterBy(inputId resource.ResId) ResMap {
	result := ResMap{}
	for id, res := range m {
		if id.Namespace() == inputId.Namespace() && id.HasSameLeftmostPrefix(inputId) {
			result[id] = res
		}
	}
	return result
}

// NewResourceSliceFromPatches returns a slice of resources given a patch path slice from a kustomization file.
func NewResourceSliceFromPatches(
	loader loader.Loader, paths []patch.PatchStrategicMerge) ([]*resource.Resource, error) {
	var result []*resource.Resource
	for _, path := range paths {
		content, err := loader.Load(string(path))
		if err != nil {
			return nil, err
		}
		res, err := newResourceSliceFromBytes(content)
		if err != nil {
			return nil, internal.Handler(err, string(path))
		}
		result = append(result, res...)
	}
	return result, nil
}

// NewResMapFromFiles returns a ResMap given a resource path slice.
func NewResMapFromFiles(loader loader.Loader, paths []string) (ResMap, error) {
	var result []ResMap
	for _, path := range paths {
		content, err := loader.Load(path)
		if err != nil {
			return nil, errors.Wrap(err, "Load from path "+path+" failed")
		}
		res, err := newResMapFromBytes(content)
		if err != nil {
			return nil, internal.Handler(err, path)
		}
		result = append(result, res)
	}
	return MergeWithoutOverride(result...)
}

// newResMapFromBytes decodes a list of objects in byte array format.
func newResMapFromBytes(b []byte) (ResMap, error) {
	resources, err := newResourceSliceFromBytes(b)
	if err != nil {
		return nil, err
	}

	result := ResMap{}
	for _, res := range resources {
		id := res.Id()
		if _, found := result[id]; found {
			return result, fmt.Errorf("GroupVersionKindName: %#v already exists b the map", id)
		}
		result[id] = res
	}
	return result, nil
}

func newResMapFromResourceSlice(resources []*resource.Resource) (ResMap, error) {
	result := ResMap{}
	for _, res := range resources {
		id := res.Id()
		if _, found := result[id]; found {
			return nil, fmt.Errorf("duplicated %#v is not allowed", id)
		}
		result[id] = res
	}
	return result, nil
}

func newResourceSliceFromBytes(in []byte) ([]*resource.Resource, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	var result []*resource.Resource
	var err error
	for err == nil || isEmptyYamlError(err) {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err == nil {
			result = append(result, resource.NewResourceFromUnstruct(out))
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

// MergeWithoutOverride combines multiple ResMap instances, failing on key collision
// and skipping nil maps. In case if all of the maps are nil, an empty ResMap is returned.
func MergeWithoutOverride(maps ...ResMap) (ResMap, error) {
	result := ResMap{}
	for _, m := range maps {
		if m == nil {
			continue
		}
		for id, res := range m {
			if _, found := result[id]; found {
				return nil, fmt.Errorf("id '%q' already used", id)
			}
			result[id] = res
		}
	}
	return result, nil
}

// MergeWithOverride combines multiple ResMap instances, allowing and sometimes
// demanding certain collisions and skipping nil maps.
// In case if all of the maps are nil, an empty ResMap is returned.
// When looping over the instances to combine them, if a resource id for resource X
// is found to be already in the combined map, then the behavior field for X
// must be BehaviorMerge or BehaviorReplace.  If X is not in the map, then it's
// behavior cannot be merge or replace.
func MergeWithOverride(maps ...ResMap) (ResMap, error) {
	result := maps[0]
	if result == nil {
		result = ResMap{}
	}
	for _, m := range maps[1:] {
		if m == nil {
			continue
		}
		for id, r := range m {
			matchedId := result.FindByGVKN(id)
			if len(matchedId) == 1 {
				id = matchedId[0]
				switch r.Behavior() {
				case resource.BehaviorReplace:
					glog.V(4).Infof("Replace %v with %v", result[id].Object, r.Object)
					r.Replace(result[id])
					result[id] = r
					result[id].SetBehavior(resource.BehaviorCreate)
				case resource.BehaviorMerge:
					glog.V(4).Infof("Merging %v with %v", result[id].Object, r.Object)
					r.Merge(result[id])
					result[id] = r
					glog.V(4).Infof("Merged object is %v", result[id].Object)
					result[id].SetBehavior(resource.BehaviorCreate)
				default:
					return nil, fmt.Errorf("id %#v exists; must merge or replace", id)
				}
			} else if len(matchedId) == 0 {
				switch r.Behavior() {
				case resource.BehaviorMerge, resource.BehaviorReplace:
					return nil, fmt.Errorf("id %#v does not exist; cannot merge or replace", id)
				default:
					result[id] = r
				}
			} else {
				return nil, fmt.Errorf("merge conflict, found multiple objects %v the Resmap %v can merge into", matchedId, id)
			}
		}
	}
	return result, nil
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
}
