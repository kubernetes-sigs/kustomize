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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kustomize/resource"
)

func TestPrefixNameRun(t *testing.T) {
	m := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: "cm1",
		}: &resource.Resource{
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
		}: &resource.Resource{
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
	expected := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: "cm1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "someprefix-cm1",
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: "cm2",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "someprefix-cm2",
					},
				},
			},
		},
	}

	npt, err := NewDefaultingNamePrefixTransformer("someprefix-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = npt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = compareMap(m, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
