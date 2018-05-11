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

package transformers

import (
	"fmt"
	"reflect"

	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/kustomize/types"
)

func compareMap(m1, m2 resource.ResourceCollection) error {
	if len(m1) != len(m2) {
		keySet1 := []types.GroupVersionKindName{}
		keySet2 := []types.GroupVersionKindName{}
		for GVKn := range m1 {
			keySet1 = append(keySet1, GVKn)
		}
		for GVKn := range m1 {
			keySet2 = append(keySet2, GVKn)
		}
		return fmt.Errorf("maps has different number of entries: %#v doesn't equals %#v", keySet1, keySet2)
	}
	for GVKn, obj1 := range m1 {
		obj2, found := m2[GVKn]
		if !found {
			return fmt.Errorf("%#v doesn't exist in %#v", GVKn, m2)
		}
		if !reflect.DeepEqual(obj1.Data, obj2.Data) {
			return fmt.Errorf("%#v doesn't match %#v", obj1.Data, obj2.Data)
		}
	}
	return nil
}
