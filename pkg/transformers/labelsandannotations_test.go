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

func TestLabelsRun(t *testing.T) {
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
						"name": "cm1",
						"labels": map[string]interface{}{
							"label-key1": "label-value1",
							"label-key2": "label-value2",
						},
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
						"labels": map[string]interface{}{
							"label-key1": "label-value1",
							"label-key2": "label-value2",
						},
					},
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"label-key1": "label-value1",
								"label-key2": "label-value2",
							},
						},
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"old-label":  "old-value",
									"label-key1": "label-value1",
									"label-key2": "label-value2",
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
						"labels": map[string]interface{}{
							"label-key1": "label-value1",
							"label-key2": "label-value2",
						},
					},
					"spec": map[string]interface{}{
						"ports": []interface{}{
							map[string]interface{}{
								"name": "port1",
								"port": "12345",
							},
						},
						"selector": map[string]interface{}{
							"label-key1": "label-value1",
							"label-key2": "label-value2",
						},
					},
				},
			},
		},
	}

	lt, err := NewDefaultingLabelsMapTransformer(map[string]string{"label-key1": "label-value1", "label-key2": "label-value2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = compareMap(m, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func makeAnnotatededConfigMap() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
				"annotations": map[string]interface{}{
					"anno-key1": "anno-value1",
					"anno-key2": "anno-value2",
				},
			},
		},
	}
}

func makeAnnotatededDeployment() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"group":      "apps",
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "deploy1",
				"annotations": map[string]interface{}{
					"anno-key1": "anno-value1",
					"anno-key2": "anno-value2",
				},
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"anno-key1": "anno-value1",
							"anno-key2": "anno-value2",
						},
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
	}
}

func makeAnnotatededService() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "svc1",
				"annotations": map[string]interface{}{
					"anno-key1": "anno-value1",
					"anno-key2": "anno-value2",
				},
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
	}
}

func TestAnnotationsRun(t *testing.T) {
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
						"name": "cm1",
						"annotations": map[string]interface{}{
							"anno-key1": "anno-value1",
							"anno-key2": "anno-value2",
						},
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
						"annotations": map[string]interface{}{
							"anno-key1": "anno-value1",
							"anno-key2": "anno-value2",
						},
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"annotations": map[string]interface{}{
									"anno-key1": "anno-value1",
									"anno-key2": "anno-value2",
								},
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
						"annotations": map[string]interface{}{
							"anno-key1": "anno-value1",
							"anno-key2": "anno-value2",
						},
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
	}
	at, err := NewDefaultingAnnotationsMapTransformer(map[string]string{"anno-key1": "anno-value1", "anno-key2": "anno-value2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = at.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = compareMap(m, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
