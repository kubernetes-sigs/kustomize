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
	"reflect"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resource"
)

// ResMap is a map from ResId to Resource.
type ResMap map[resid.ResId]*resource.Resource

// FindByGVKN find the matched ResIds by Group/Version/Kind and Name
func (m ResMap) FindByGVKN(inputId resid.ResId) []resid.ResId {
	var result []resid.ResId
	for id := range m {
		if id.GvknEquals(inputId) {
			result = append(result, id)
		}
	}
	return result
}

// DemandOneMatchForId find the matched resource by Group/Version/Kind and Name
func (m ResMap) DemandOneMatchForId(inputId resid.ResId) (*resource.Resource, bool) {
	result := m.FindByGVKN(inputId)
	if len(result) == 1 {
		return m[result[0]], true
	}
	return nil, false
}

// EncodeAsYaml encodes a ResMap to YAML; encoded objects separated by `---`.
func (m ResMap) EncodeAsYaml() ([]byte, error) {
	var ids []resid.ResId
	for id := range m {
		ids = append(ids, id)
	}
	sort.Sort(IdSlice(ids))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, id := range ids {
		obj := m[id]
		out, err := yaml.Marshal(obj.Map())
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
		var keySet1 []resid.ResId
		var keySet2 []resid.ResId
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
func (m ResMap) DeepCopy(rf *resource.Factory) ResMap {
	mcopy := make(ResMap)
	for id, obj := range m {
		mcopy[id] = rf.FromKunstructured(obj.Copy())
		mcopy[id].SetBehavior(obj.Behavior())
	}
	return mcopy
}

// FilterBy returns a ResMap containing ResIds with the same namespace and nameprefix
// with the inputId
func (m ResMap) FilterBy(inputId resid.ResId) ResMap {
	result := ResMap{}
	for id, res := range m {
		if id.Namespace() == inputId.Namespace() && id.HasSameLeftmostPrefix(inputId) {
			result[id] = res
		}
	}
	return result
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
// A collision would be demanded, say, when a generated ConfigMap has the
// "replace" option in its generation instructions, meaning it is supposed
// to replace something from the raw resources list.
// If all of the maps are nil, an empty ResMap is returned.
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
				case ifc.BehaviorReplace:
					glog.V(4).Infof(
						"Replace %v with %v", result[id].Map(), r.Map())
					r.Replace(result[id])
					result[id] = r
					result[id].SetBehavior(ifc.BehaviorCreate)
				case ifc.BehaviorMerge:
					glog.V(4).Infof(
						"Merging %v with %v", result[id].Map(), r.Map())
					r.Merge(result[id])
					result[id] = r
					glog.V(4).Infof(
						"Merged object is %v", result[id].Map())
					result[id].SetBehavior(ifc.BehaviorCreate)
				default:
					return nil, fmt.Errorf("id %#v exists; must merge or replace", id)
				}
			} else if len(matchedId) == 0 {
				switch r.Behavior() {
				case ifc.BehaviorMerge, ifc.BehaviorReplace:
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
