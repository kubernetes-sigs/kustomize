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

func TestNameReferenceRun(t *testing.T) {
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
						"name": "someprefix-cm1-somehash",
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
						"name": "someprefix-secret1-somehash",
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
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.7.9",
										"env": []interface{}{
											map[string]interface{}{
												"name": "CM_FOO",
												"valueFrom": map[string]interface{}{
													"configMapKeyRef": map[string]interface{}{
														"name": "cm1",
														"key":  "somekey",
													},
												},
											},
											map[string]interface{}{
												"name": "SECRET_FOO",
												"valueFrom": map[string]interface{}{
													"secretKeyRef": map[string]interface{}{
														"name": "secret1",
														"key":  "somekey",
													},
												},
											},
										},
										"envFrom": []interface{}{
											map[string]interface{}{
												"configMapRef": map[string]interface{}{
													"name": "cm1",
													"key":  "somekey",
												},
											},
											map[string]interface{}{
												"secretRef": map[string]interface{}{
													"name": "secret1",
													"key":  "somekey",
												},
											},
										},
									},
								},
								"volumes": map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": "cm1",
									},
									"secret": map[string]interface{}{
										"secretName": "secret1",
									},
								},
							},
						},
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
						"name": "someprefix-cm1-somehash",
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
						"name": "someprefix-secret1-somehash",
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
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.7.9",
										"env": []interface{}{
											map[string]interface{}{
												"name": "CM_FOO",
												"valueFrom": map[string]interface{}{
													"configMapKeyRef": map[string]interface{}{
														"name": "someprefix-cm1-somehash",
														"key":  "somekey",
													},
												},
											},
											map[string]interface{}{
												"name": "SECRET_FOO",
												"valueFrom": map[string]interface{}{
													"secretKeyRef": map[string]interface{}{
														"name": "someprefix-secret1-somehash",
														"key":  "somekey",
													},
												},
											},
										},
										"envFrom": []interface{}{
											map[string]interface{}{
												"configMapRef": map[string]interface{}{
													"name": "someprefix-cm1-somehash",
													"key":  "somekey",
												},
											},
											map[string]interface{}{
												"secretRef": map[string]interface{}{
													"name": "someprefix-secret1-somehash",
													"key":  "somekey",
												},
											},
										},
									},
								},
								"volumes": map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": "someprefix-cm1-somehash",
									},
									"secret": map[string]interface{}{
										"secretName": "someprefix-secret1-somehash",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	nrt, err := NewDefaultingNameReferenceTransformer()
	err = nrt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = compareMap(m, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
