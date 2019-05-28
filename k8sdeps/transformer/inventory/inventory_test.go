// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
	"sigs.k8s.io/kustomize/pkg/validators"
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

func TestInventoryTransformer(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	ldr := loader.NewFileLoaderAtCwd(validators.MakeFakeValidator(), fs.MakeFakeFS())

	// hash is derived based on all keys in the Inventory
	// It is added to annotations as
	//   kustomize.config.k8s.io/InventoryHash: hash
	// When seeing the same annotation, prune binary assumes no
	// clean up is needed
	hash := "h44788gt7g"

	// inventory is the derived json string for an Inventory object
	// It is added to annotations as
	// kustomize.config.k8s.io/Inventory: inventory
	inventory := "{\"current\":{\"apps_v1_Deployment|~X|deploy1\":null,\"~G_v1_ConfigMap|~X|cm1\":[{\"group\":\"apps\",\"version\":\"v1\",\"kind\":\"Deployment\",\"name\":\"deploy1\"}],\"~G_v1_Secret|~X|secret1\":[{\"group\":\"apps\",\"version\":\"v1\",\"kind\":\"Deployment\",\"name\":\"deploy1\"}]}}" // nolint

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
					"kustomize.config.k8s.io/Inventory":     inventory,
					"kustomize.config.k8s.io/InventoryHash": hash,
				},
			},
		})
	expected := resmap.ResMap{
		resid.NewResIdWithPrefixNamespace(cmap, "pruneCM", "", "default"): pruneMap,
	}

	p := &types.Inventory{
		Type: "ConfigMap",
		ConfigMap: types.NameArgs{
			Name:      "pruneCM",
			Namespace: "default",
		},
	}
	objs := makeResMap()

	// include the original resmap; only return the ConfigMap for pruning
	tran := NewInventoryTransformer(p, ldr, "default", types.GarbageCollect)
	tran.Transform(objs)

	if !reflect.DeepEqual(objs, expected) {
		err := expected.ErrorIfNotEqual(objs)
		t.Fatalf("actual doesn't match expected: %v", err)
	}

	objs = makeResMap()
	expected = objs.DeepCopy(rf)
	expected[resid.NewResIdWithPrefixNamespace(cmap, "pruneCM", "", "default")] = pruneMap
	// append the ConfigMap for pruning to the original resmap
	tran = NewInventoryTransformer(p, ldr, "default", types.GarbageIgnore)
	tran.Transform(objs)

	if !reflect.DeepEqual(objs, expected) {
		err := expected.ErrorIfNotEqual(objs)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
