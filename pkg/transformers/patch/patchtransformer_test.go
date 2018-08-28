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

package patch

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
)

func TestMultipleTypePatches(t *testing.T) {
	base := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
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
			}),
	}
	patchSM := []*resource.Resource{
		resource.NewResourceFromMap(map[string]interface{}{
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
		),
		resource.NewResourceFromMap(map[string]interface{}{
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
		),
	}

	patchJ6 := map[resource.ResId][]byte{
		resource.NewResId(deploy, "deploy1"): []byte(`[
             {"op": "add", "path": "/spec/replica", "value": "3"},
             {"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:latest"}
]`),
	}
	expected := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
				},
				"spec": map[string]interface{}{
					"replica": "3",
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
			}),
	}
	lt, err := NewPatchTransformer(patchSM, patchJ6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(base, expected) {
		err = expected.ErrorIfNotEqual(base)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestMultipleTypePatchesWithConflict(t *testing.T) {
	base := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
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
			}),
	}
	patchSM := []*resource.Resource{
		resource.NewResourceFromMap(map[string]interface{}{
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
		),
		resource.NewResourceFromMap(map[string]interface{}{
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
		),
	}

	patchJ6 := map[resource.ResId][]byte{
		resource.NewResId(deploy, "deploy1"): []byte(`[
             {"op": "add", "path": "/spec/replica", "value": "3"},
             {"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:newest"}
]`),
	}
	lt, err := NewPatchTransformer(patchSM, patchJ6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err == nil {
		t.Fatal("expected conflict")
	}
	if !strings.Contains(err.Error(), "There is conflict between different types of patches.") {
		t.Fatalf("expected conflict, but got: %v", err)
	}
}

func TestOneTypePatches(t *testing.T) {
	base := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
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
			}),
	}
	patchSM := []*resource.Resource{
		resource.NewResourceFromMap(map[string]interface{}{
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
		),
		resource.NewResourceFromMap(map[string]interface{}{
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
		),
	}
	expected := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
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
			}),
	}
	patchJ6 := map[resource.ResId][]byte{}
	lt, err := NewPatchTransformer(patchSM, patchJ6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(base, expected) {
		err = expected.ErrorIfNotEqual(base)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
