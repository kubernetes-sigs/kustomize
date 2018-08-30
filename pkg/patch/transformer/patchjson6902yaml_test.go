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
	"testing"

	yamlpatch "github.com/krishicks/yaml-patch"
	"github.com/kubernetes-sigs/kustomize/pkg/patch"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var deploy = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}

func TestJsonPatchYAMLTransformer_Transform(t *testing.T) {
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
	}

	target := patch.Target{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
		Name:    "deploy1",
	}

	var image, replica, command interface{}
	image = "my-nginx"
	replica = "3"
	command = []string{"arg1", "arg2", "arg3"}
	patch := yamlpatch.Patch{
		{
			Op:    "replace",
			Path:  "/spec/template/spec/containers/0/name",
			Value: yamlpatch.NewNode(&image),
		},
		{
			Op:    "add",
			Path:  "/spec/replica",
			Value: yamlpatch.NewNode(&replica),
		},
		{
			Op:    "add",
			Path:  "/spec/template/spec/containers/0/command",
			Value: yamlpatch.NewNode(&command),
		},
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
	jpt, err := NewPatchJson6902YAMLTransformer(&target, patch)
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
