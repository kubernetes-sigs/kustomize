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

func TestNameHashTransformer(t *testing.T) {
	objs := resource.ResourceCollection{
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
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"group":      "apps",
					"apiVersion": "v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "deploy1",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"old-label": "old-value",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.7.9",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "Service"},
			Name: "svc1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name": "svc1",
					},
					"spec": map[string]interface{}{
						"ports": []interface{}{
							map[string]interface{}{
								"name": "port1",
								"port": "12345",
							},
						},
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "Secret"},
			Name: "secret1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"name": "secret1",
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
						"name": "cm1-m462kdfb68",
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"group":      "apps",
					"apiVersion": "v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "deploy1",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"old-label": "old-value",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.7.9",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "Service"},
			Name: "svc1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name": "svc1",
					},
					"spec": map[string]interface{}{
						"ports": []interface{}{
							map[string]interface{}{
								"name": "port1",
								"port": "12345",
							},
						},
					},
				},
			},
		},
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "Secret"},
			Name: "secret1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"name": "secret1-7kc45hd5f7",
					},
				},
			},
		},
	}

	tran := NewNameHashTransformer()
	tran.Transform(objs)

	if !reflect.DeepEqual(objs, expected) {
		err := compareMap(objs, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
