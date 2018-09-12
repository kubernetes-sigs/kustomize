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

package resmap

import (
	"reflect"
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/pkg/resource"
)

func TestLess(t *testing.T) {
	ids := IdSlice{
		resource.NewResId(schema.GroupVersionKind{Kind: "ConfigMap"}, "cm"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Pod"}, "pod"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Namespace"}, "ns1"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Namespace"}, "ns2"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Role"}, "ro"),
		resource.NewResId(schema.GroupVersionKind{Kind: "RoleBinding"}, "rb"),
		resource.NewResId(schema.GroupVersionKind{Kind: "CustomResourceDefinition"}, "crd"),
		resource.NewResId(schema.GroupVersionKind{Kind: "ServiceAccount"}, "sa"),
	}
	expected := IdSlice{
		resource.NewResId(schema.GroupVersionKind{Kind: "Namespace"}, "ns1"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Namespace"}, "ns2"),
		resource.NewResId(schema.GroupVersionKind{Kind: "CustomResourceDefinition"}, "crd"),
		resource.NewResId(schema.GroupVersionKind{Kind: "ServiceAccount"}, "sa"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Role"}, "ro"),
		resource.NewResId(schema.GroupVersionKind{Kind: "RoleBinding"}, "rb"),
		resource.NewResId(schema.GroupVersionKind{Kind: "ConfigMap"}, "cm"),
		resource.NewResId(schema.GroupVersionKind{Kind: "Pod"}, "pod"),
	}
	sort.Sort(ids)
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("expected %+v but got %+v", expected, ids)
	}
}
