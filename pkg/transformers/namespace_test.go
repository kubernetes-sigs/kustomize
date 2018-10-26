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

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

func TestNamespaceRun(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	m := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resid.NewResId(cmap, "cm2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "cm2",
					"namespace": "foo",
				},
			}),
		resid.NewResId(ns, "ns1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "ns1",
				},
			}),
		resid.NewResId(sa, "default"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "default",
					"namespace": "system",
				},
			}),
		resid.NewResId(sa, "service-account"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "service-account",
					"namespace": "system",
				},
			}),
		resid.NewResId(crb, "crb"): rf.FromMap(
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
		resid.NewResId(crd, "crd"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1beta1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd",
				},
			}),
	}
	expected := resmap.ResMap{
		resid.NewResIdWithPrefixNamespace(ns, "ns1", "", ""): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "ns1",
				},
			}),
		resid.NewResIdWithPrefixNamespace(cmap, "cm1", "", "test"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "cm1",
					"namespace": "test",
				},
			}),
		resid.NewResIdWithPrefixNamespace(cmap, "cm2", "", "test"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "cm2",
					"namespace": "test",
				},
			}),
		resid.NewResIdWithPrefixNamespace(sa, "default", "", "test"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "default",
					"namespace": "test",
				},
			}),
		resid.NewResIdWithPrefixNamespace(sa, "service-account", "", "test"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "service-account",
					"namespace": "test",
				},
			}),
		resid.NewResId(crb, "crb"): rf.FromMap(
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
		resid.NewResId(crd, "crd"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1beta1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd",
				},
			}),
	}

	nst := NewNamespaceTransformer("test", defaultTransformerConfig.NameSpace)
	err := nst.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestNamespaceRunForClusterLevelKind(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	m := resmap.ResMap{
		resid.NewResId(ns, "ns1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "ns1",
				},
			}),
		resid.NewResId(crd, "crd1"): rf.FromMap(
			map[string]interface{}{
				"kind": "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd1",
				},
			}),
		resid.NewResId(pv, "pv1"): rf.FromMap(
			map[string]interface{}{
				"kind": "PersistentVolume",
				"metadata": map[string]interface{}{
					"name": "pv1",
				},
			}),
		resid.NewResId(cr, "cr1"): rf.FromMap(
			map[string]interface{}{
				"kind": "ClusterRole",
				"metadata": map[string]interface{}{
					"name": "cr1",
				},
			}),
		resid.NewResId(crb, "crb1"): rf.FromMap(
			map[string]interface{}{
				"kind": "ClusterRoleBinding",
				"metadata": map[string]interface{}{
					"name": "crb1",
				},
				"subjects": []interface{}{},
			}),
	}

	expected := m.DeepCopy(rf)

	nst := NewNamespaceTransformer("test", defaultTransformerConfig.NameSpace)

	err := nst.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
