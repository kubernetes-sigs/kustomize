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

package resmap

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resource"
)

var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
var statefulset = gvk.Gvk{Group: "apps", Version: "v1", Kind: "StatefulSet"}
var rf = resource.NewFactory(
	kunstruct.NewKunstructuredFactoryImpl())
var rmF = NewFactory(rf)

func TestEncodeAsYaml(t *testing.T) {
	encoded := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`)
	input := ResMap{
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
					"name": "cm2",
				},
			}),
	}
	out, err := input.EncodeAsYaml()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, encoded) {
		t.Fatalf("%s doesn't match expected %s", out, encoded)
	}
}

func TestDemandOneMatchForId(t *testing.T) {
	rm1 := ResMap{
		resid.NewResIdWithPrefixNamespace(cmap, "cm1", "prefix1", "ns1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resid.NewResIdWithPrefixNamespace(cmap, "cm2", "prefix1", "ns1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm2",
				},
			}),
	}

	_, ok := rm1.DemandOneMatchForId(resid.NewResIdWithPrefixNamespace(cmap, "cm2", "prefix1", "ns1"))
	if !ok {
		t.Fatal("Expected single map entry but got none")
	}

	// confirm that ns and prefix are not included in match
	_, ok = rm1.DemandOneMatchForId(resid.NewResIdWithPrefixNamespace(cmap, "cm2", "prefix", "ns"))
	if !ok {
		t.Fatal("Expected single map entry but got none")
	}

	// confirm that name is matched correctly
	result, ok := rm1.DemandOneMatchForId(resid.NewResIdWithPrefixNamespace(cmap, "cm3", "prefix1", "ns1"))
	if ok {
		t.Fatalf("Expected no map entries but got %v", result)
	}

	cmap2 := gvk.Gvk{Version: "v2", Kind: "ConfigMap"}

	// confirm that gvk is matched correctly
	result, ok = rm1.DemandOneMatchForId(resid.NewResIdWithPrefixNamespace(cmap2, "cm2", "prefix1", "ns1"))
	if ok {
		t.Fatalf("Expected no map entries but got %v", result)
	}

}

func TestFilterBy(t *testing.T) {
	tests := map[string]struct {
		resMap   ResMap
		filter   resid.ResId
		expected ResMap
	}{
		"different namespace": {
			resMap: ResMap{resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix", "namespace1"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
			filter:   resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix", "namespace2"),
			expected: ResMap{},
		},
		"different prefix": {
			resMap: ResMap{resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix1", "suffix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
			filter:   resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix2", "suffix", "namespace"),
			expected: ResMap{},
		},
		"different suffix": {
			resMap: ResMap{resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix1", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
			filter:   resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix2", "namespace"),
			expected: ResMap{},
		},
		"same namespace, same prefix": {
			resMap: ResMap{resid.NewResIdWithPrefixNamespace(cmap, "config-map1", "prefix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map1",
					},
				}),
			},
			filter: resid.NewResIdWithPrefixNamespace(cmap, "config-map2", "prefix", "namespace"),
			expected: ResMap{resid.NewResIdWithPrefixNamespace(cmap, "config-map1", "prefix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map1",
					},
				}),
			},
		},
		"same namespace, same suffix": {
			resMap: ResMap{resid.NewResIdWithSuffixNamespace(cmap, "config-map1", "suffix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map1",
					},
				}),
			},
			filter: resid.NewResIdWithSuffixNamespace(cmap, "config-map2", "suffix", "namespace"),
			expected: ResMap{resid.NewResIdWithSuffixNamespace(cmap, "config-map1", "suffix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map1",
					},
				}),
			},
		},
		"same namespace, same prefix, same suffix": {
			resMap: ResMap{resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map1", "prefix", "suffix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
			filter: resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map2", "prefix", "suffix", "namespace"),
			expected: ResMap{resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map1", "prefix", "suffix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
		},
		"filter by cluster-level Gvk": {
			resMap: ResMap{resid.NewResIdWithPrefixNamespace(cmap, "config-map", "prefix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
			filter: resid.NewResId(gvk.Gvk{Kind: "ClusterRoleBinding"}, "cluster-role-binding"),
			expected: ResMap{resid.NewResIdWithPrefixNamespace(cmap, "config-map", "prefix", "namespace"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "config-map",
					},
				}),
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			got := test.resMap.FilterBy(test.filter)
			if !reflect.DeepEqual(test.expected, got) {
				t.Fatalf("Expected %v but got back %v", test.expected, got)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	rm1 := ResMap{
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
					"name": "cm2",
				},
			}),
	}

	rm2 := rm1.DeepCopy(rf)

	if &rm1 == &rm2 {
		t.Fatal("DeepCopy returned a reference to itself instead of a copy")
	}

	if !reflect.DeepEqual(rm1, rm2) {
		t.Fatalf("%v doesn't equal it's deep copy %v", rm1, rm2)
	}
}

func TestErrorIfNotEqual(t *testing.T) {

	rm1 := ResMap{
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
					"name": "cm2",
				},
			}),
	}

	err := rm1.ErrorIfNotEqual(rm1)
	if err != nil {
		t.Fatalf("%v should equal itself %v", rm1, err)
	}

	rm2 := ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
	}

	// test the different number of keys path
	err = rm1.ErrorIfNotEqual(rm2)
	if err == nil {
		t.Fatalf("%v should not equal %v %v", rm1, rm2, err)
	}

	rm3 := ResMap{
		resid.NewResId(cmap, "cm2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
	}

	// test the different key values path
	err = rm2.ErrorIfNotEqual(rm3)
	if err == nil {
		t.Fatalf("%v should not equal %v %v", rm1, rm2, err)
	}

	rm4 := ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm3",
				},
			}),
	}

	// test the deepcopy path
	err = rm2.ErrorIfNotEqual(rm4)
	if err == nil {
		t.Fatalf("%v should not equal %v %v", rm1, rm2, err)
	}

}

func TestMergeWithoutOverride(t *testing.T) {
	input1 := ResMap{
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "foo-deploy1",
				},
			}),
	}
	input2 := ResMap{
		resid.NewResId(statefulset, "stateful1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "StatefulSet",
				"metadata": map[string]interface{}{
					"name": "bar-stateful",
				},
			}),
	}
	input := []ResMap{input1, input2}
	expected := ResMap{
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "foo-deploy1",
				},
			}),
		resid.NewResId(statefulset, "stateful1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "StatefulSet",
				"metadata": map[string]interface{}{
					"name": "bar-stateful",
				},
			}),
	}
	merged, err := MergeWithErrorOnIdCollision(input...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged, expected)
	}
	input3 := []ResMap{merged, nil}
	merged1, err := MergeWithErrorOnIdCollision(input3...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged1, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged1, expected)
	}
	input4 := []ResMap{nil, merged}
	merged2, err := MergeWithErrorOnIdCollision(input4...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged2, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged2, expected)
	}
}

func generateMergeFixtures(b ifc.GenerationBehavior) []ResMap {

	input1 := ResMap{
		resid.NewResId(cmap, "cmap"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cmap",
				},
				"data": map[string]interface{}{
					"a": "x",
					"b": "y",
				},
			}),
	}
	input2 := ResMap{
		resid.NewResId(cmap, "cmap"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cmap",
				},
				"data": map[string]interface{}{
					"a": "u",
					"b": "v",
					"c": "w",
				},
			}),
	}
	input1[resid.NewResId(cmap, "cmap")].SetBehavior(ifc.BehaviorCreate)
	input2[resid.NewResId(cmap, "cmap")].SetBehavior(b)
	return []ResMap{input1, input2}
}

func TestMergeWithOverride(t *testing.T) {
	expected := ResMap{
		resid.NewResId(cmap, "cmap"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{},
					"labels":      map[string]interface{}{},
					"name":        "cmap",
				},
				"data": map[string]interface{}{
					"a": "u",
					"b": "v",
					"c": "w",
				},
			}),
	}
	expected[resid.NewResId(cmap, "cmap")].SetBehavior(ifc.BehaviorCreate)
	merged, err := MergeWithOverride(generateMergeFixtures(ifc.BehaviorMerge)...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged, expected)
	}
	input3 := []ResMap{merged, nil}
	merged1, err := MergeWithOverride(input3...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged1, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged1, expected)
	}
	input4 := []ResMap{nil, merged}
	merged2, err := MergeWithOverride(input4...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged2, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged2, expected)
	}

	inputs := generateMergeFixtures(ifc.BehaviorReplace)
	replaced, err := MergeWithOverride(inputs...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedReplaced := inputs[1]
	if !reflect.DeepEqual(replaced, expectedReplaced) {
		t.Fatalf("%#v doesn't equal expected %#v", replaced, expectedReplaced)
	}

	_, err = MergeWithOverride(generateMergeFixtures(ifc.BehaviorUnspecified)...)
	if err == nil {
		t.Fatal("Merging with GenerationBehavior BehaviorUnspecified should return an error but does not")
	}
}
