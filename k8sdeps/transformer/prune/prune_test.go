/*
Copyright 2019 The Kubernetes Authors.

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

package prune

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

var secret = gvk.Gvk{Version: "v1", Kind: "Secret"}
var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

func makeResMap() resmap.ResMap {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	objs := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resid.NewResId(secret, "secret1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "secret1",
				},
			}),
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
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
									"env": []interface{}{
										map[string]interface{}{
											"name": "CM_FOO",
											"valueFrom": map[string]interface{}{
												"configMapKeyRef": map[string]interface{}{
													"name": "cm1",
													"key":  "somekey",
												},
											},
										},
									},
									"envFrom": []interface{}{
										map[string]interface{}{
											"configMapRef": map[string]interface{}{
												"name": "cm1",
												"key":  "somekey",
											},
										},
										map[string]interface{}{
											"secretRef": map[string]interface{}{
												"name": "secret1",
												"key":  "somekey",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
	}
	objs[resid.NewResId(cmap, "cm1")].AppendRefBy(resid.NewResId(deploy, "deploy1"))
	objs[resid.NewResId(secret, "secret1")].AppendRefBy(resid.NewResId(deploy, "deploy1"))
	return objs
}

func TestPruneTransformer(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())

	// hash is derived based on all keys in the ConfigMap data field.
	// It is added to annotations as
	//   current: hash
	// When seeing the same annotation, prune binary assumes no
	// clean up is needed
	hash := "k777d7h45b"
	//  This is the root or inventory object which tracks all
	// the applied resources - this is the thing we expect the transformer to create.
	pruneMap := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "pruneCM",
				"namespace": "default",
				"annotations": map[string]interface{}{
					"current": hash,
				},
			},
			"data": map[string]interface{}{
				"_ConfigMap__cm1":                             hash,
				"_Secret__secret1":                            hash,
				"apps_Deployment__deploy1":                    hash,
				"_ConfigMap__cm1---apps_Deployment__deploy1":  hash,
				"_Secret__secret1---apps_Deployment__deploy1": hash,
			},
		})
	expected := resmap.ResMap{
		resid.NewResIdWithPrefixNamespace(cmap, "pruneCM", "", "default"): pruneMap,
	}

	p := &types.Prune{
		Type: "alphaConfigMap",
		AlphaConfigMap: types.NameArgs{
			Name:      "pruneCM",
			Namespace: "default",
		},
	}
	objs := makeResMap()

	// include the original resmap; only return the ConfigMap for pruning
	tran := NewPruneTransformer(p, "default", false)
	tran.Transform(objs)

	if !reflect.DeepEqual(objs, expected) {
		err := expected.ErrorIfNotEqual(objs)
		t.Fatalf("actual doesn't match expected: %v", err)
	}

	objs = makeResMap()
	expected = objs.DeepCopy(rf)
	expected[resid.NewResIdWithPrefixNamespace(cmap, "pruneCM", "", "default")] = pruneMap
	// append the ConfigMap for pruning to the original resmap
	tran = NewPruneTransformer(p, "default", true)
	tran.Transform(objs)

	if !reflect.DeepEqual(objs, expected) {
		err := expected.ErrorIfNotEqual(objs)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
