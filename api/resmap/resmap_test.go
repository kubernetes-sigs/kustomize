// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmap_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	. "sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/testutils/resmaptest"
	"sigs.k8s.io/kustomize/api/types"
)

var rf = resource.NewFactory(
	kunstruct.NewKunstructuredFactoryImpl())
var rmF = NewFactory(rf, nil)

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

// Maintain the class invariant that no two
// resources can have the same CurId().
func TestAppendRejectsDuplicateResId(t *testing.T) {
	w := New()
	if err := w.Append(makeCm(1)); err != nil {
		t.Fatalf("append error: %v", err)
	}
	err := w.Append(makeCm(1))
	if err == nil {
		t.Fatalf("expected append error")
	}
	if !strings.Contains(
		err.Error(),
		"may not add resource with an already registered id") {
		t.Fatalf("unexpected error: %v", err)
	}
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
	doRemove(t, w1, makeCm(1).OrgId())
	doRemove(t, w1, makeCm(3).OrgId())
	doRemove(t, w1, makeCm(5).OrgId())
	doRemove(t, w1, makeCm(7).OrgId())

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
	err := w.Remove(r.OrgId())
	if err == nil {
		t.Fatalf("expected error")
	}
	err = w.Append(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = w.Remove(r.OrgId())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = w.Remove(r.OrgId())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReplace(t *testing.T) {
	cm5 := makeCm(5)
	cm700 := makeCm(700)
	otherCm5 := makeCm(5)

	w := New()
	doAppend(t, w, makeCm(1))
	doAppend(t, w, makeCm(2))
	doAppend(t, w, makeCm(3))
	doAppend(t, w, makeCm(4))
	doAppend(t, w, cm5)
	doAppend(t, w, makeCm(6))
	doAppend(t, w, makeCm(7))

	oldSize := w.Size()
	_, err := w.Replace(otherCm5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Size() != oldSize {
		t.Fatalf("unexpected size %d", w.Size())
	}
	if r, err := w.GetByCurrentId(cm5.OrgId()); err != nil || r != otherCm5 {
		t.Fatalf("unexpected result r=%s, err=%v", r.CurId(), err)
	}
	if err := w.Append(cm5); err == nil {
		t.Fatalf("expected id already there error")
	}
	if err := w.Remove(cm5.OrgId()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := w.Append(cm700); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := w.Append(cm5); err != nil {
		t.Fatalf("unexpected err: %v", err)
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
	input := resmaptest_test.NewRmBuilder(t, rf).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			},
		}).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm2",
			},
		}).ResMap()
	out, err := input.AsYaml()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, encoded) {
		t.Fatalf("%s doesn't match expected %s", out, encoded)
	}
}

func TestGetMatchingResourcesByCurrentId(t *testing.T) {
	r1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "alice",
			},
		})
	r2 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "bob",
			},
		})
	r3 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "bob",
				"namespace": "happy",
			},
		})
	r4 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "charlie",
				"namespace": "happy",
			},
		})
	r5 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "charlie",
				"namespace": "happy",
			},
		})

	m := resmaptest_test.NewRmBuilder(t, rf).
		AddR(r1).AddR(r2).AddR(r3).AddR(r4).AddR(r5).ResMap()

	result := m.GetMatchingResourcesByCurrentId(
		resid.NewResId(cmap, "alice").GvknEquals)
	if len(result) != 1 {
		t.Fatalf("Expected single map entry but got %v", result)
	}
	result = m.GetMatchingResourcesByCurrentId(
		resid.NewResId(cmap, "bob").GvknEquals)
	if len(result) != 2 {
		t.Fatalf("Expected two, got %v", result)
	}
	result = m.GetMatchingResourcesByCurrentId(
		resid.NewResIdWithNamespace(cmap, "bob", "system").GvknEquals)
	if len(result) != 2 {
		t.Fatalf("Expected two but got %v", result)
	}
	result = m.GetMatchingResourcesByCurrentId(
		resid.NewResIdWithNamespace(cmap, "bob", "happy").Equals)
	if len(result) != 1 {
		t.Fatalf("Expected single map entry but got %v", result)
	}
	result = m.GetMatchingResourcesByCurrentId(
		resid.NewResId(cmap, "charlie").GvknEquals)
	if len(result) != 1 {
		t.Fatalf("Expected single map entry but got %v", result)
	}

	// nolint:goconst
	tests := []struct {
		name    string
		matcher IdMatcher
		count   int
	}{
		{
			"match everything",
			func(resid.ResId) bool { return true },
			5,
		},
		{
			"match nothing",
			func(resid.ResId) bool { return false },
			0,
		},
		{
			"name is alice",
			func(x resid.ResId) bool { return x.Name == "alice" },
			1,
		},
		{
			"name is charlie",
			func(x resid.ResId) bool { return x.Name == "charlie" },
			2,
		},
		{
			"name is bob",
			func(x resid.ResId) bool { return x.Name == "bob" },
			2,
		},
		{
			"happy namespace",
			func(x resid.ResId) bool {
				return x.Namespace == "happy"
			},
			3,
		},
		{
			"happy deployment",
			func(x resid.ResId) bool {
				return x.Namespace == "happy" &&
					x.Gvk.Kind == "Deployment"
			},
			1,
		},
		{
			"happy ConfigMap",
			func(x resid.ResId) bool {
				return x.Namespace == "happy" &&
					x.Gvk.Kind == "ConfigMap"
			},
			2,
		},
	}
	for _, tst := range tests {
		result := m.GetMatchingResourcesByCurrentId(tst.matcher)
		if len(result) != tst.count {
			t.Fatalf("test '%s';  actual: %d, expected: %d",
				tst.name, len(result), tst.count)
		}
	}
}

func TestSubsetThatCouldBeReferencedByResource(t *testing.T) {
	r1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "alice",
			},
		})
	r2 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "bob",
			},
		})
	r3 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "bob",
				"namespace": "happy",
			},
		})
	r4 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "charlie",
				"namespace": "happy",
			},
		})
	r5 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "charlie",
				"namespace": "happy",
			},
		})
	r5.AddNamePrefix("little-")
	r6 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "domino",
				"namespace": "happy",
			},
		})
	r6.AddNamePrefix("little-")
	r7 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ClusterRoleBinding",
			"metadata": map[string]interface{}{
				"name": "meh",
			},
		})

	tests := map[string]struct {
		filter   *resource.Resource
		expected ResMap
	}{
		"default namespace 1": {
			filter: r2,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(r1).AddR(r2).AddR(r7).ResMap(),
		},
		"default namespace 2": {
			filter: r1,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(r1).AddR(r2).AddR(r7).ResMap(),
		},
		"happy namespace no prefix": {
			filter: r3,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(r3).AddR(r4).AddR(r5).AddR(r6).AddR(r7).ResMap(),
		},
		"happy namespace with prefix": {
			filter: r5,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(r3).AddR(r4).AddR(r5).AddR(r6).AddR(r7).ResMap(),
		},
		"cluster level": {
			filter: r7,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(r1).AddR(r2).AddR(r3).AddR(r4).AddR(r5).AddR(r6).AddR(r7).ResMap(),
		},
	}
	m := resmaptest_test.NewRmBuilder(t, rf).
		AddR(r1).AddR(r2).AddR(r3).AddR(r4).AddR(r5).AddR(r6).AddR(r7).ResMap()
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			got := m.SubsetThatCouldBeReferencedByResource(test.filter)
			err := test.expected.ErrorIfNotEqualLists(got)
			if err != nil {
				test.expected.Debug("expected")
				got.Debug("actual")
				t.Fatalf("Expected match")
			}
		})
	}
}

func addPfxSfx(r *resource.Resource, prefixes []string, suffixes []string) {
	for _, pfx := range prefixes {
		r.AddNamePrefix(pfx)
	}

	for _, sfx := range suffixes {
		r.AddNameSuffix(sfx)
	}
}

func TestSubsetThatCouldBeReferencedByResourceMultiLevel(t *testing.T) {
	// Simulates ConfigMap and Deployment defined at level 1
	// No prefix nor suffix added at that level
	cm1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "level1",
			},
		})
	addPfxSfx(cm1, []string{""}, []string{""})
	dep1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "level1",
			},
		})
	addPfxSfx(dep1, []string{""}, []string{""})

	// Simulates ConfigMap and Deployment defined at level 1
	// and prefix added in level 2 of kustomization
	cm2p := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "level2p",
			},
		})
	addPfxSfx(cm2p, []string{"", "level2p-"}, []string{"", ""})

	dep2p := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "level2p",
			},
		})
	addPfxSfx(dep2p, []string{"", "level2p-"}, []string{"", ""})

	// Simulates ConfigMap and Deployment defined at level 1
	// and suffix added in level 2 of kustomization
	cm2s := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "level2s",
			},
		})
	addPfxSfx(cm2s, []string{"", ""}, []string{"", "-level2s"})

	dep2s := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "level2s",
			},
		})
	addPfxSfx(dep2s, []string{"", ""}, []string{"", "-level2s"})

	// Simulates ConfigMap and Deployment defined at level 1,
	// prefix added in levels 2 and 3 of kustomization.
	cm3e := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "level3e",
			},
		})
	addPfxSfx(cm3e, []string{"", "level2p-", "level3e-"}, []string{"", "", ""})

	dep3e := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "level3e",
			},
		})
	addPfxSfx(dep3e, []string{"", "level2p-", "level3e-"}, []string{"", "", ""})

	// Simulates Deployment defined at level 1, ConfigMap defined at level 2,
	// prefix added in levels 2 and 3 of kustomization.
	// This reproduce issue 1440.
	cm3i := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "level3i",
			},
		})
	addPfxSfx(cm3i, []string{"level2p-", "level3i-"}, []string{"", ""})

	dep3i := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "level3i",
			},
		})
	addPfxSfx(dep3i, []string{"", "level2p-", "level3i-"}, []string{"", "", ""})

	tests := map[string]struct {
		filter   *resource.Resource
		expected ResMap
	}{
		"level1": {
			filter: dep1,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(cm1).AddR(dep1).ResMap(),
		},
		"level2p": {
			filter: dep2p,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(cm2p).AddR(dep2p).ResMap(),
		},
		"level2s": {
			filter: dep2s,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(cm2s).AddR(dep2s).ResMap(),
		},
		"level3p": {
			filter: dep3e,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(cm3e).AddR(dep3e).ResMap(),
		},
		"level3i": {
			filter: dep3i,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				AddR(cm3i).AddR(dep3i).ResMap(),
		},
	}
	m := resmaptest_test.NewRmBuilder(t, rf).
		AddR(cm1).AddR(dep1).AddR(cm2s).AddR(dep2s).AddR(cm2p).AddR(dep2p).AddR(cm3e).AddR(dep3e).AddR(cm3i).AddR(dep3i).ResMap()
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			got := m.SubsetThatCouldBeReferencedByResource(test.filter)
			err := test.expected.ErrorIfNotEqualLists(got)
			if err != nil {
				test.expected.Debug("expected")
				got.Debug("actual")
				t.Fatalf("Expected match")
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	rm1 := resmaptest_test.NewRmBuilder(t, rf).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			},
		}).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm2",
			},
		}).ResMap()

	rm2 := rm1.DeepCopy()

	if &rm1 == &rm2 {
		t.Fatal("DeepCopy returned a reference to itself instead of a copy")
	}
	err := rm1.ErrorIfNotEqualLists(rm1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorIfNotEqualSets(t *testing.T) {
	r1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			},
		})
	r2 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm2",
			},
		})
	r3 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "cm2",
				"namespace": "system",
			},
		})

	m1 := resmaptest_test.NewRmBuilder(t, rf).AddR(r1).AddR(r2).AddR(r3).ResMap()
	if err := m1.ErrorIfNotEqualSets(m1); err != nil {
		t.Fatalf("object should equal itself %v", err)
	}

	m2 := resmaptest_test.NewRmBuilder(t, rf).AddR(r1).ResMap()
	if err := m1.ErrorIfNotEqualSets(m2); err == nil {
		t.Fatalf("%v should not equal %v %v", m1, m2, err)
	}

	m3 := resmaptest_test.NewRmBuilder(t, rf).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			}}).ResMap()
	if err := m2.ErrorIfNotEqualSets(m3); err != nil {
		t.Fatalf("%v should equal %v %v", m2, m3, err)
	}

	m4 := resmaptest_test.NewRmBuilder(t, rf).AddR(r1).AddR(r2).AddR(r3).ResMap()
	if err := m1.ErrorIfNotEqualSets(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
	}

	m4 = resmaptest_test.NewRmBuilder(t, rf).AddR(r3).AddR(r1).AddR(r2).ResMap()
	if err := m1.ErrorIfNotEqualSets(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
	}

	m4 = m1.ShallowCopy()
	if err := m1.ErrorIfNotEqualSets(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
	}
	m4 = m1.DeepCopy()
	if err := m1.ErrorIfNotEqualSets(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
	}
}

func TestErrorIfNotEqualLists(t *testing.T) {
	r1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			},
		})
	r2 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm2",
			},
		})
	r3 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "cm2",
				"namespace": "system",
			},
		})

	m1 := resmaptest_test.NewRmBuilder(t, rf).AddR(r1).AddR(r2).AddR(r3).ResMap()
	if err := m1.ErrorIfNotEqualLists(m1); err != nil {
		t.Fatalf("object should equal itself %v", err)
	}

	m2 := resmaptest_test.NewRmBuilder(t, rf).AddR(r1).ResMap()
	if err := m1.ErrorIfNotEqualLists(m2); err == nil {
		t.Fatalf("%v should not equal %v %v", m1, m2, err)
	}

	m3 := resmaptest_test.NewRmBuilder(t, rf).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			}}).ResMap()
	if err := m2.ErrorIfNotEqualLists(m3); err != nil {
		t.Fatalf("%v should equal %v %v", m2, m3, err)
	}

	m4 := resmaptest_test.NewRmBuilder(t, rf).AddR(r1).AddR(r2).AddR(r3).ResMap()
	if err := m1.ErrorIfNotEqualLists(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
	}

	m4 = resmaptest_test.NewRmBuilder(t, rf).AddR(r3).AddR(r1).AddR(r2).ResMap()
	if err := m1.ErrorIfNotEqualLists(m4); err == nil {
		t.Fatalf("expected inequality between %v and %v, %v", m1, m4, err)
	}

	m4 = m1.ShallowCopy()
	if err := m1.ErrorIfNotEqualLists(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
	}
	m4 = m1.DeepCopy()
	if err := m1.ErrorIfNotEqualLists(m4); err != nil {
		t.Fatalf("expected equality between %v and %v, %v", m1, m4, err)
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
	if err := expected.Append(r1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := expected.Append(r2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := input1.AppendAll(input2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := expected.ErrorIfNotEqualLists(input1); err != nil {
		input1.Debug("1")
		expected.Debug("ex")
		t.Fatalf("%#v doesn't equal expected %#v", input1, expected)
	}
	if err := input1.AppendAll(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := expected.ErrorIfNotEqualLists(input1); err != nil {
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
	if err := expected.ErrorIfNotEqualLists(w); err != nil {
		t.Fatal(err)
	}
	w = makeMap1()
	if err := w.AbsorbAll(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.ErrorIfNotEqualLists(makeMap1()); err != nil {
		t.Fatal(err)
	}
	w = makeMap1()
	w2 := makeMap2(types.BehaviorReplace)
	if err := w.AbsorbAll(w2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w2.ErrorIfNotEqualLists(w); err != nil {
		t.Fatal(err)
	}
	w = makeMap1()
	w2 = makeMap2(types.BehaviorUnspecified)
	err := w.AbsorbAll(w2)
	if err == nil {
		t.Fatalf("expected error with unspecified behavior")
	}
}
