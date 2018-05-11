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
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kustomize/resource"
)

func TestOverlayRun(t *testing.T) {
	base := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
										"image": "nginx",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	overlay := []*resource.Resource{
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "deploy1",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"another-label": "foo",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:latest",
										"env": []interface{}{
											map[string]interface{}{
												"name":  "SOMEENV",
												"value": "BAR",
											},
										},
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
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "deploy1",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"old-label":     "old-value",
									"another-label": "foo",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:latest",
										"env": []interface{}{
											map[string]interface{}{
												"name":  "SOMEENV",
												"value": "BAR",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	lt, err := NewOverlayTransformer(overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(base, expected) {
		err = compareMap(base, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestMultiplePatches(t *testing.T) {
	base := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
										"image": "nginx",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	overlay := []*resource.Resource{
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
										"image": "nginx:latest",
										"env": []interface{}{
											map[string]interface{}{
												"name":  "SOMEENV",
												"value": "BAR",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "deploy1",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name": "nginx",
										"env": []interface{}{
											map[string]interface{}{
												"name":  "ANOTHERENV",
												"value": "HELLO",
											},
										},
									},
									map[string]interface{}{
										"name":  "busybox",
										"image": "busybox",
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
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
										"image": "nginx:latest",
										"env": []interface{}{
											map[string]interface{}{
												"name":  "ANOTHERENV",
												"value": "HELLO",
											},
											map[string]interface{}{
												"name":  "SOMEENV",
												"value": "BAR",
											},
										},
									},
									map[string]interface{}{
										"name":  "busybox",
										"image": "busybox",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	lt, err := NewOverlayTransformer(overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(base, expected) {
		err = compareMap(base, expected)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestMultiplePatchesWithConflict(t *testing.T) {
	base := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "deploy1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
										"image": "nginx",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	overlay := []*resource.Resource{
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
										"image": "nginx:latest",
										"env": []interface{}{
											map[string]interface{}{
												"name":  "SOMEENV",
												"value": "BAR",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
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
									},
								},
							},
						},
					},
				},
			},
		},
	}

	lt, err := NewOverlayTransformer(overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err == nil {
		t.Fatalf("did not get expected error")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("expected error to contain %q but get %v", "conflict", err)
	}
}

func TestNoSchemaOverlayRun(t *testing.T) {
	base := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Foo"},
			Name: "my-foo",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"A": "X",
							"B": "Y",
						},
					},
				},
			},
		},
	}
	overlay := []*resource.Resource{
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"B": nil,
							"C": "Z",
						},
					},
				},
			},
		},
	}
	expected := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Foo"},
			Name: "my-foo",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"A": "X",
							"C": "Z",
						},
					},
				},
			},
		},
	}

	lt, err := NewOverlayTransformer(overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err = compareMap(base, expected); err != nil {
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestNoSchemaMultiplePatches(t *testing.T) {
	base := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Foo"},
			Name: "my-foo",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"A": "X",
							"B": "Y",
						},
					},
				},
			},
		},
	}
	overlay := []*resource.Resource{
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"B": nil,
							"C": "Z",
						},
					},
				},
			},
		},
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"C": "Z",
							"D": "W",
						},
						"baz": map[string]interface{}{
							"hello": "world",
						},
					},
				},
			},
		},
	}
	expected := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Foo"},
			Name: "my-foo",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"A": "X",
							"C": "Z",
							"D": "W",
						},
						"baz": map[string]interface{}{
							"hello": "world",
						},
					},
				},
			},
		},
	}

	lt, err := NewOverlayTransformer(overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err = compareMap(base, expected); err != nil {
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestNoSchemaMultiplePatchesWithConflict(t *testing.T) {
	base := resource.ResourceCollection{
		{
			GVK:  schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Foo"},
			Name: "my-foo",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"A": "X",
							"B": "Y",
						},
					},
				},
			},
		},
	}
	overlay := []*resource.Resource{
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"B": nil,
							"C": "Z",
						},
					},
				},
			},
		},
		{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "Foo",
					"metadata": map[string]interface{}{
						"name": "my-foo",
					},
					"spec": map[string]interface{}{
						"bar": map[string]interface{}{
							"C": "NOT_Z",
						},
					},
				},
			},
		},
	}

	lt, err := NewOverlayTransformer(overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err == nil {
		t.Fatalf("did not get expected error")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("expected error to contain %q but get %v", "conflict", err)
	}
}
