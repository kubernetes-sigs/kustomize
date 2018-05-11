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
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kustomize/types"
)

func TestDecodeToResourceCollection(t *testing.T) {
	encoded := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`)
	expected := ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: "cm1",
		}: &Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "cm1",
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: "cm2",
		}: &Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "cm2",
					},
				},
			},
		},
	}
	m, err := decodeToResourceCollection(encoded)
	fmt.Printf("%v\n", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		t.Fatalf("%#v doesn't match expected %#v", m, expected)
	}
}

func TestMerge(t *testing.T) {
	input1 := ResourceCollection{
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "foo-deploy1",
					},
				},
			},
		},
	}
	input2 := ResourceCollection{
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
			Name: "stateful1",
		}: &Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "StatefulSet",
					"metadata": map[string]interface{}{
						"name": "bar-stateful",
					},
				},
			},
		},
	}
	input := []ResourceCollection{input1, input2}
	expected := ResourceCollection{
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "foo-deploy1",
					},
				},
			},
		},
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
			Name: "stateful1",
		}: &Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "StatefulSet",
					"metadata": map[string]interface{}{
						"name": "bar-stateful",
					},
				},
			},
		},
	}
	merged, err := Merge(input...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged, expected)
	}
}
