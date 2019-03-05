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

package transformer

import (
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

func TestJsonPatchJSONTransformer_Transform(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	id := resid.NewResId(deploy, "deploy1")
	base := resmap.ResMap{
		id: rf.FromMap(
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
	}

	operations := []byte(`[
        {"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
        {"op": "add", "path": "/spec/replica", "value": "3"},
        {"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)

	expected := resmap.ResMap{
		id: rf.FromMap(
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
								"old-label": "old-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
									"name":  "my-nginx",
									"command": []interface{}{
										"arg1",
										"arg2",
										"arg3",
									},
								},
							},
						},
					},
				},
			}),
	}
	jpt, err := newPatchJson6902JSONTransformer(id, operations)
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

func TestJsonPatchJSONTransformer_UnHappyTransform(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	id := resid.NewResId(deploy, "deploy1")
	base := resmap.ResMap{
		id: rf.FromMap(
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
	}

	operations := []byte(`[
        {"op": "add", "path": "/spec/template/spec/containers/0/command/", "value": ["arg1", "arg2", "arg3"]}
]`)

	jpt, err := newPatchJson6902JSONTransformer(id, operations)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	err = jpt.Transform(base)
	if err == nil {
		t.Fatalf("expected error didn't happen")
	}
	if !strings.HasPrefix(err.Error(), "failed to apply json patch") || !strings.Contains(err.Error(), string(operations)) {
		t.Fatalf("expected error didn't happen, but got %v", err)
	}
}
