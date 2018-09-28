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

	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformerconfig"
)

func TestNamespaceRun(t *testing.T) {
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
					"name":      "cm2",
					"namespace": "foo",
				},
			}),
		resource.NewResId(ns, "ns1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "ns1",
				},
			}),
		resource.NewResId(sa, "default"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "default",
					"namespace": "system",
				},
			}),
		resource.NewResId(sa, "service-account"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "service-account",
					"namespace": "system",
				},
			}),
		resource.NewResId(crb, "crb"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRoleBinding",
				"metadata": map[string]interface{}{
					"name": "manager-rolebinding",
				},
				"subjects": []interface{}{
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "default",
						"namespace": "system",
					},
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "service-account",
						"namespace": "system",
					},
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "another",
						"namespace": "random",
					},
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
		resource.NewResIdWithPrefixNamespace(ns, "ns1", "", ""): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "ns1",
				},
			}),
		resource.NewResIdWithPrefixNamespace(cmap, "cm1", "", "test"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "cm1",
					"namespace": "test",
				},
			}),
		resource.NewResIdWithPrefixNamespace(cmap, "cm2", "", "test"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "cm2",
					"namespace": "test",
				},
			}),
		resource.NewResIdWithPrefixNamespace(sa, "default", "", "test"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "default",
					"namespace": "test",
				},
			}),
		resource.NewResIdWithPrefixNamespace(sa, "service-account", "", "test"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "service-account",
					"namespace": "test",
				},
			}),
		resource.NewResId(crb, "crb"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRoleBinding",
				"metadata": map[string]interface{}{
					"name": "manager-rolebinding",
				},
				"subjects": []interface{}{
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "default",
						"namespace": "test",
					},
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "service-account",
						"namespace": "test",
					},
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "another",
						"namespace": "random",
					},
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

	nst := NewNamespaceTransformer("test", transformerconfig.MakeDefaultTransformerConfig().NameSpace)
	err := nst.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
