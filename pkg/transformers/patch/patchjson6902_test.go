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
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var deploy = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
var foo = schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Foo"}

func TestJsonPatchTransformer_Transform(t *testing.T) {
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
			}),
		resource.NewResIdWithPrefixNamespace(deploy, "deploy2", "", "test"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy2",
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
			}),
	}
	patches := map[resource.ResId][]byte{
		resource.NewResId(deploy, "deploy1"): []byte(`[
             {"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
             {"op": "add", "path": "/spec/replica", "value": "3"},
             {"op": "remove", "path": "/spec/template/metadata/labels/old-label"},
             {"op": "add", "path": "/spec/template/metadata/labels/new-label", "value": "new-value"}
]`),
		resource.NewResIdWithPrefixNamespace(deploy, "deploy2", "", "test"): []byte(`[
             {"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
             {"op": "add", "path": "/spec/replica", "value": "3"},
             {"op": "remove", "path": "/spec/template/metadata/labels/old-label"},
             {"op": "add", "path": "/spec/template/metadata/labels/new-label", "value": "new-value"}
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
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"new-label": "new-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "my-nginx",
									"image": "nginx",
								},
							},
						},
					},
				},
			}),
		resource.NewResIdWithPrefixNamespace(deploy, "deploy2", "", "test"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy2",
				},
				"spec": map[string]interface{}{
					"replica": "3",
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"new-label": "new-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "my-nginx",
									"image": "nginx",
								},
							},
						},
					},
				},
			}),
	}
	jpt, err := NewPatchJson6902Transformer(patches)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	err = jpt.Transform(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(base, expected) {
		err = expected.ErrorIfNotEqual(base)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
