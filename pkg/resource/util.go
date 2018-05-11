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
	"fmt"
	"io"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/kustomize/constants"
)

// decode decodes a list of objects in byte array format
func decode(in []byte) ([]*Resource, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	resources := []*Resource{}

	var err error
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			break
		}
		resources = append(resources, &Resource{Data: &out})
	}
	if err != io.EOF {
		return nil, err
	}
	return resources, nil
}

// decodeToResourceCollection decodes a list of objects in byte array format.
// it will return a ResourceCollection.
func decodeToResourceCollection(in []byte) (ResourceCollection, error) {
	resources, err := decode(in)
	if err != nil {
		return nil, err
	}

	into := ResourceCollection{}
	for _, res := range resources {
		gvkn := res.GVKN()
		if _, found := into[gvkn]; found {
			return into, fmt.Errorf("GroupVersionKindName: %#v already exists in the map", gvkn)
		}
		into[gvkn] = res
	}
	return into, nil
}

func resourceCollectionFromResources(resources []*Resource) (ResourceCollection, error) {
	out := ResourceCollection{}
	for _, res := range resources {
		gvkn := res.GVKN()
		if _, found := out[gvkn]; found {
			return nil, fmt.Errorf("duplicated %#v is not allowed", gvkn)
		}
		out[gvkn] = res
	}
	return out, nil
}

// Merge will merge all of the entries in the slice of ResourceCollection.
func Merge(rcs ...ResourceCollection) (ResourceCollection, error) {
	all := ResourceCollection{}
	for _, rc := range rcs {
		for gvkn, obj := range rc {
			if _, found := all[gvkn]; found {
				return nil, fmt.Errorf("there is already an entry: %q", gvkn)
			}
			all[gvkn] = obj
		}
	}

	return all, nil
}

// MergeWithOverride will merge all of the entries in the slice of ResourceCollection with Override
// If there is already an entry with the same GVKN exists, different actions are performed according to value of Behavior field
// 'create': create a new one;
// 'replace': replace the data only; keep the labels and annotations
// 'merge': merge the data; keep the labels and annotations
func MergeWithOverride(rcs ...ResourceCollection) (ResourceCollection, error) {
	all := ResourceCollection{}

	for _, rc := range rcs {
		for gvkn, obj := range rc {
			if _, found := all[gvkn]; found {
				switch obj.Behavior {
				case "", constants.CreateBehavior:
					return nil, fmt.Errorf("Create an existing gvkn %#v is not allowed", gvkn)
				case constants.ReplaceBehavior:
					glog.V(4).Infof("Replace object %v by %v", all[gvkn].Data.Object, obj.Data.Object)
					obj.replace(all[gvkn])
					all[gvkn] = obj
				case constants.MergeBehavior:
					glog.V(4).Infof("Merge object %v with %v", all[gvkn].Data.Object, obj.Data.Object)
					obj.merge(all[gvkn])
					all[gvkn] = obj
					glog.V(4).Infof("The merged object is %v", all[gvkn].Data.Object)
				default:
					return nil, fmt.Errorf("The behavior of %#v must be one of merge and replace since it already exists in the base", gvkn)
				}
			} else {
				switch obj.Behavior {
				case "", constants.CreateBehavior:
					all[gvkn] = obj
				case constants.MergeBehavior, constants.ReplaceBehavior:
					return nil, fmt.Errorf("No merge or replace is allowed for non existing gvkn %#v", gvkn)
				default:
					return nil, fmt.Errorf("The behavior of %#v must be create since it doesn't exist", gvkn)
				}
			}
		}
	}
	return all, nil
}
func (r *Resource) replace(other *Resource) {
	r.Data.SetLabels(mergeMap(other.Data.GetLabels(), r.Data.GetLabels()))
	r.Data.SetAnnotations(mergeMap(other.Data.GetAnnotations(), r.Data.GetAnnotations()))
	r.Data.SetName(other.Data.GetName())
}

func (r *Resource) merge(other *Resource) {
	r.replace(other)
	mergeConfigmap(r.Data.Object, other.Data.Object, r.Data.Object)
}

func mergeMap(maps ...map[string]string) map[string]string {
	mergedMap := map[string]string{}
	for _, m := range maps {
		for key, value := range m {
			mergedMap[key] = value
		}
	}
	return mergedMap
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
