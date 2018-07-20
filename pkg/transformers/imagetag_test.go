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

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestImageTagTransformer(t *testing.T) {
	m := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
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
								},
								map[string]interface{}{
									"name":  "nginx2",
									"image": "my-nginx:1.8.0",
								},
							},
						},
					},
				},
			}),
		resource.NewResId(schema.GroupVersionKind{Kind: "randomeKind"}, "random"): resource.NewResourceFromMap(
			map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx1",
									"image": "nginx",
								},
								map[string]interface{}{
									"name":  "nginx2",
									"image": "my-nginx:random",
								},
							},
						},
					},
				},
				"spec2": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "ngin3",
									"image": "nginx:v1",
								},
								map[string]interface{}{
									"name":  "nginx4",
									"image": "my-nginx:latest",
								},
							},
						},
					},
				},
			}),
	}
	expected := resmap.ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
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
									"image": "nginx:v2",
								},
								map[string]interface{}{
									"name":  "nginx2",
									"image": "my-nginx:previous",
								},
							},
						},
					},
				},
			}),
		resource.NewResId(schema.GroupVersionKind{Kind: "randomeKind"}, "random"): resource.NewResourceFromMap(
			map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx1",
									"image": "nginx:v2",
								},
								map[string]interface{}{
									"name":  "nginx2",
									"image": "my-nginx:previous",
								},
							},
						},
					},
				},
				"spec2": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "ngin3",
									"image": "nginx:v2",
								},
								map[string]interface{}{
									"name":  "nginx4",
									"image": "my-nginx:previous",
								},
							},
						},
					},
				},
			}),
	}

	it, err := NewImageTagTransformer([]types.ImageTag{
		{Name: "nginx", NewTag: "v2"},
		{Name: "my-nginx", NewTag: "previous"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = it.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
