// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmap_test

import (
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	. "sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
var rf = resource.NewFactory(
	kunstruct.NewKunstructuredFactoryImpl())
var rmF = NewFactory(rf)

func doAppend(t *testing.T, w ResMap, r *resource.Resource) {
	err := w.Append(r)
	if err != nil {
		t.Fatalf("append error: %v", err)
	}
}
func doRemove(t *testing.T, w ResMap, id resid.ResId) {
	err := w.Remove(id)
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}
}

// Make a resource with a predictable name.
func makeCm(i int) *resource.Resource {
	return rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("cm%03d", i),
			},
		})
}

func TestAppendRemove(t *testing.T) {
	w1 := New()
	doAppend(t, w1, makeCm(1))
	doAppend(t, w1, makeCm(2))
	doAppend(t, w1, makeCm(3))
	doAppend(t, w1, makeCm(4))
	doAppend(t, w1, makeCm(5))
	doAppend(t, w1, makeCm(6))
	doAppend(t, w1, makeCm(7))
	doRemove(t, w1, makeCm(1).Id())
	doRemove(t, w1, makeCm(3).Id())
	doRemove(t, w1, makeCm(5).Id())
	doRemove(t, w1, makeCm(7).Id())

	w2 := New()
	doAppend(t, w2, makeCm(2))
	doAppend(t, w2, makeCm(4))
	doAppend(t, w2, makeCm(6))
	if !reflect.DeepEqual(w1, w1) {
		w1.Debug("w1")
		w2.Debug("w2")
		t.Fatalf("mismatch")
	}

	err := w2.Append(makeCm(6))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRemove(t *testing.T) {
	w := New()
	r := makeCm(1)
	err := w.Remove(r.Id())
	if err == nil {
		t.Fatalf("expected error")
	}
	err = w.Append(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = w.Remove(r.Id())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = w.Remove(r.Id())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReplaceResource(t *testing.T) {
	cm5 := makeCm(5)
	cm700 := makeCm(700)
	cm888 := makeCm(888)

	w := New()
	doAppend(t, w, makeCm(1))
	doAppend(t, w, makeCm(2))
	doAppend(t, w, makeCm(3))
	doAppend(t, w, makeCm(4))
	doAppend(t, w, cm5)
	doAppend(t, w, makeCm(6))
	doAppend(t, w, makeCm(7))

	oldSize := w.Size()
	err := w.ReplaceResource(cm5.Id(), cm700)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Size() != oldSize {
		t.Fatalf("unexpected size %d", w.Size())
	}
	if w.GetById(cm5.Id()) != cm700 {
		t.Fatalf("unexpected result")
	}
	if err := w.Append(cm5); err == nil {
		t.Fatalf("expected id already there error")
	}
	if err := w.AppendWithId(cm888.Id(), cm5); err != nil {
		// Okay to add with some unused Id.
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Append(cm700); err == nil {
		t.Fatalf("expected resource already there error")
	}
	if err := w.Remove(cm5.Id()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := w.Append(cm700); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := w.Append(cm5); err == nil {
		t.Fatalf("expected err; object is still there under id 888")
	}
	if err := w.Remove(cm888.Id()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := w.Append(cm5); err != nil {
		t.Fatalf("unexpected err; %v", err)
	}
}

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
	input := New()
	input.Append(rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			},
		}))
	input.Append(rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm2",
			},
		}))
	out, err := input.AsYaml(Identity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, encoded) {
		t.Fatalf("%s doesn't match expected %s", out, encoded)
	}
}

func TestDemandOneGvknMatchForId(t *testing.T) {
	rm1 := FromMap(map[resid.ResId]*resource.Resource{
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
	})

	result := rm1.GetMatchingIds(
		resid.NewResIdWithPrefixNamespace(cmap, "cm2", "prefix1", "ns1").GvknEquals)
	if len(result) != 1 {
		t.Fatalf("Expected single map entry but got %v", result)
	}

	// confirm that ns and prefix are not included in match
	result = rm1.GetMatchingIds(
		resid.NewResIdWithPrefixNamespace(cmap, "cm2", "prefix", "ns").GvknEquals)
	if len(result) != 1 {
		t.Fatalf("Expected single map entry but got %v", result)
	}

	// confirm that name is matched correctly
	result = rm1.GetMatchingIds(
		resid.NewResIdWithPrefixNamespace(cmap, "cm3", "prefix1", "ns1").GvknEquals)
	if len(result) > 0 {
		t.Fatalf("Expected no map entries but got %v", result)
	}

	cmap2 := gvk.Gvk{Version: "v2", Kind: "ConfigMap"}

	// confirm that gvk is matched correctly
	result = rm1.GetMatchingIds(
		resid.NewResIdWithPrefixNamespace(cmap2, "cm2", "prefix1", "ns1").GvknEquals)
	if len(result) > 0 {
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
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix", "namespace1"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
			filter:   resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix", "namespace2"),
			expected: New(),
		},
		"different prefix": {
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix1", "suffix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
			filter:   resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix2", "suffix", "namespace"),
			expected: New(),
		},
		"different suffix": {
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix1", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
			filter:   resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map", "prefix", "suffix2", "namespace"),
			expected: New(),
		},
		"same namespace, same prefix": {
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixNamespace(cmap, "config-map1", "prefix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map1",
						},
					}),
			}),
			filter: resid.NewResIdWithPrefixNamespace(cmap, "config-map2", "prefix", "namespace"),
			expected: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixNamespace(cmap, "config-map1", "prefix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map1",
						},
					}),
			}),
		},
		"same namespace, same suffix": {
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithSuffixNamespace(cmap, "config-map1", "suffix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map1",
						},
					}),
			}),
			filter: resid.NewResIdWithSuffixNamespace(cmap, "config-map2", "suffix", "namespace"),
			expected: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithSuffixNamespace(cmap, "config-map1", "suffix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map1",
						},
					}),
			}),
		},
		"same namespace, same prefix, same suffix": {
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map1", "prefix", "suffix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
			filter: resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map2", "prefix", "suffix", "namespace"),
			expected: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixSuffixNamespace(cmap, "config-map1", "prefix", "suffix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
		},
		"filter by cluster-level Gvk": {
			resMap: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixNamespace(cmap, "config-map", "prefix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
			filter: resid.NewResId(gvk.Gvk{Kind: "ClusterRoleBinding"}, "cluster-role-binding"),
			expected: FromMap(map[resid.ResId]*resource.Resource{
				resid.NewResIdWithPrefixNamespace(cmap, "config-map", "prefix", "namespace"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "config-map",
						},
					}),
			}),
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			got := test.resMap.ResourcesThatCouldReference(test.filter)
			err := test.expected.ErrorIfNotEqual(got)
			if err != nil {
				t.Fatalf("Expected %v but got back %v", test.expected, got)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	rm1 := FromMap(map[resid.ResId]*resource.Resource{
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
	})

	rm2 := rm1.DeepCopy()

	if &rm1 == &rm2 {
		t.Fatal("DeepCopy returned a reference to itself instead of a copy")
	}
	err := rm1.ErrorIfNotEqual(rm1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetMatchingIds(t *testing.T) {

	m := FromMap(map[resid.ResId]*resource.Resource{
		resid.NewResId(
			gvk.Gvk{Kind: "vegetable"},
			"bedlam"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "whatever1",
				},
			}),
		resid.NewResId(
			gvk.Gvk{Group: "g1", Version: "v1", Kind: "vegetable"},
			"domino"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "whatever2",
				},
			}),
		resid.NewResIdWithPrefixNamespace(
			gvk.Gvk{Kind: "vegetable"},
			"peter", "p", "happy"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "whatever3",
				},
			}),
		resid.NewResIdWithPrefixNamespace(
			gvk.Gvk{Version: "v1", Kind: "fruit"},
			"shatterstar", "p", "happy"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "whatever4",
				},
			}),
	})

	tests := []struct {
		name    string
		matcher IdMatcher
		count   int
	}{
		{
			"match everything",
			func(resid.ResId) bool { return true },
			4,
		},
		{
			"match nothing",
			func(resid.ResId) bool { return false },
			0,
		},
		{
			"name is peter",
			func(x resid.ResId) bool { return x.Name() == "peter" },
			1,
		},
		{
			"happy vegetable",
			func(x resid.ResId) bool {
				return x.Namespace() == "happy" &&
					x.Gvk().Kind == "vegetable"
			},
			1,
		},
	}
	for _, tst := range tests {
		result := m.GetMatchingIds(tst.matcher)
		if len(result) != tst.count {
			t.Fatalf("test '%s';  actual: %d, expected: %d",
				tst.name, len(result), tst.count)
		}
	}
}

func TestErrorIfNotEqual(t *testing.T) {
	rm1 := FromMap(map[resid.ResId]*resource.Resource{
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
	})

	err := rm1.ErrorIfNotEqual(rm1)
	if err != nil {
		t.Fatalf("%v should equal itself %v", rm1, err)
	}

	rm2 := FromMap(map[resid.ResId]*resource.Resource{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
	})

	// test the different number of keys path
	err = rm1.ErrorIfNotEqual(rm2)
	if err == nil {
		t.Fatalf("%v should not equal %v %v", rm1, rm2, err)
	}

	rm3 := FromMap(map[resid.ResId]*resource.Resource{
		resid.NewResId(cmap, "cm2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
	})

	// test the different key values path
	err = rm2.ErrorIfNotEqual(rm3)
	if err == nil {
		t.Fatalf("%v should not equal %v %v", rm1, rm2, err)
	}

	rm4 := FromMap(map[resid.ResId]*resource.Resource{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm3",
				},
			}),
	})

	// test the deepcopy path
	err = rm2.ErrorIfNotEqual(rm4)
	if err == nil {
		t.Fatalf("%v should not equal %v %v", rm1, rm2, err)
	}
}

func TestAppendAll(t *testing.T) {
	r1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "foo-deploy1",
			},
		})
	input1 := rmF.FromResource(r1)
	r2 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "StatefulSet",
			"metadata": map[string]interface{}{
				"name": "bar-stateful",
			},
		})
	input2 := rmF.FromResource(r2)

	expected := New()
	expected.Append(r1)
	expected.Append(r2)

	if err := input1.AppendAll(input2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := expected.ErrorIfNotEqual(input1); err != nil {
		input1.Debug("1")
		expected.Debug("ex")
		t.Fatalf("%#v doesn't equal expected %#v", input1, expected)
	}
	if err := input1.AppendAll(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := expected.ErrorIfNotEqual(input1); err != nil {
		t.Fatalf("%#v doesn't equal expected %#v", input1, expected)
	}
}

func makeMap1() ResMap {
	return rmF.FromResource(rf.FromMapAndOption(
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
		}, &types.GeneratorArgs{
			Behavior: "create",
		}, nil))
}

func makeMap2(b types.GenerationBehavior) ResMap {
	return rmF.FromResource(rf.FromMapAndOption(
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
		}, &types.GeneratorArgs{
			Behavior: b.String(),
		}, nil))
}

func TestAbsorbAll(t *testing.T) {
	expected := rmF.FromResource(rf.FromMapAndOption(
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
		}, &types.GeneratorArgs{
			Behavior: "create",
		}, nil))
	w := makeMap1()
	if err := w.AbsorbAll(makeMap2(types.BehaviorMerge)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := expected.ErrorIfNotEqual(w); err != nil {
		t.Fatal(err)
	}
	w = makeMap1()
	if err := w.AbsorbAll(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.ErrorIfNotEqual(makeMap1()); err != nil {
		t.Fatal(err)
	}
	w = makeMap1()
	w2 := makeMap2(types.BehaviorReplace)
	if err := w.AbsorbAll(w2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w2.ErrorIfNotEqual(w); err != nil {
		t.Fatal(err)
	}
	w = makeMap1()
	w2 = makeMap2(types.BehaviorUnspecified)
	err := w.AbsorbAll(w2)
	if err == nil {
		t.Fatalf("expected error with unspecified behavior")
	}
}
