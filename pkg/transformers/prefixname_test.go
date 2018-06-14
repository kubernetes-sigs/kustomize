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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPrefixNameRun(t *testing.T) {
	m := resmap.ResMap{
		resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resource.NewResId(cmap, "cm2"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm2",
				},
			}),
		resource.NewResId(crd, "crd"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1beta1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd",
				},
			}),
	}
	expected := resmap.ResMap{
		resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm1",
				},
			}),
		resource.NewResId(cmap, "cm2"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm2",
				},
			}),
		resource.NewResId(crd, "crd"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1beta1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd",
				},
			}),
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
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestAddPrefixPathConfigs(t *testing.T) {
	expected := len(defaultNamePrefixPathConfigs) + 1

	pathConfigs := []PathConfig{
		{
			GroupVersionKind:   &schema.GroupVersionKind{Group: "GroupA", Kind: "KindB"},
			Path:               []string{"path", "to", "a", "field"},
			CreateIfNotPresent: true,
		},
	}
	AddPrefixPathConfigs(pathConfigs...)
	if len(defaultNamePrefixPathConfigs) != expected {
		t.Fatalf("actual %v doesn't match expected: %v", len(defaultNamePrefixPathConfigs), expected)
	}
}
