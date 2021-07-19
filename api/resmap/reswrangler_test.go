// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmap_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/labels"
	"sigs.k8s.io/kustomize/api/provider"
	. "sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	resmaptest_test "sigs.k8s.io/kustomize/api/testutils/resmaptest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var depProvider = provider.NewDefaultDepProvider()
var rf = depProvider.GetResourceFactory()
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
	cmap := resid.NewGvk("", "v1", "ConfigMap")

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

func TestGetMatchingResourcesByAnyId(t *testing.T) {
	r1 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "new-alice",
				"annotations": map[string]interface{}{
					"config.kubernetes.io/previousKinds":      "ConfigMap",
					"config.kubernetes.io/previousNames":      "alice",
					"config.kubernetes.io/previousNamespaces": "default",
				},
			},
		})
	r2 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "new-bob",
				"annotations": map[string]interface{}{
					"config.kubernetes.io/previousKinds":      "ConfigMap,ConfigMap",
					"config.kubernetes.io/previousNames":      "bob,bob2",
					"config.kubernetes.io/previousNamespaces": "default,default",
				},
			},
		})
	r3 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "new-bob",
				"namespace": "new-happy",
				"annotations": map[string]interface{}{
					"config.kubernetes.io/previousKinds":      "ConfigMap",
					"config.kubernetes.io/previousNames":      "bob",
					"config.kubernetes.io/previousNamespaces": "happy",
				},
			},
		})
	r4 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "charlie",
				"namespace": "happy",
				"annotations": map[string]interface{}{
					"config.kubernetes.io/previousKinds":      "ConfigMap",
					"config.kubernetes.io/previousNames":      "charlie",
					"config.kubernetes.io/previousNamespaces": "default",
				},
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
		result := m.GetMatchingResourcesByAnyId(tst.matcher)
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
			"apiVersion": "apps/v1",
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
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "domino",
				"namespace": "happy",
			},
		})
	r6.AddNamePrefix("little-")
	r7 := rf.FromMap(
		map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
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

	m3 := resmaptest_test.NewRmBuilder(t, rf).AddR(r2).ResMap()
	if err := m2.ErrorIfNotEqualSets(m3); err == nil {
		t.Fatalf("%v should not equal %v %v", m2, m3, err)
	}

	m3 = resmaptest_test.NewRmBuilder(t, rf).Add(
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
		}))
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
		}))
}

func TestAbsorbAll(t *testing.T) {
	metadata := map[string]interface{}{
		"name": "cmap",
	}
	expected := rmF.FromResource(rf.FromMapAndOption(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "ConfigMap",
			"metadata":   metadata,
			"data": map[string]interface{}{
				"a": "u",
				"b": "v",
				"c": "w",
			},
		},
		&types.GeneratorArgs{
			Behavior: "create",
		}))
	w := makeMap1()
	assert.NoError(t, w.AbsorbAll(makeMap2(types.BehaviorMerge)))
	expected.RemoveBuildAnnotations()
	w.RemoveBuildAnnotations()
	assert.NoError(t, expected.ErrorIfNotEqualLists(w))
	w = makeMap1()
	assert.NoError(t, w.AbsorbAll(nil))
	assert.NoError(t, w.ErrorIfNotEqualLists(makeMap1()))

	w = makeMap1()
	w2 := makeMap2(types.BehaviorReplace)
	assert.NoError(t, w.AbsorbAll(w2))
	w2.RemoveBuildAnnotations()
	assert.NoError(t, w2.ErrorIfNotEqualLists(w))
	w = makeMap1()
	w2 = makeMap2(types.BehaviorUnspecified)
	err := w.AbsorbAll(w2)
	assert.Error(t, err)
	assert.True(
		t, strings.Contains(err.Error(), "behavior must be merge or replace"))
}

func TestToRNodeSlice(t *testing.T) {
	input := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: namespace-reader
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - watch
  - list
`
	rm, err := rmF.NewResMapFromBytes([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := bytes.NewBufferString("")
	for i, n := range rm.ToRNodeSlice() {
		if i != 0 {
			b.WriteString("---\n")
		}
		s, err := n.String()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b.WriteString(s)
	}

	if !reflect.DeepEqual(input, b.String()) {
		t.Fatalf("actual doesn't match expected.\nActual:\n%s\n===\nExpected:\n%s\n",
			b.String(), input)
	}
}

func TestApplySmPatch_General(t *testing.T) {
	const (
		myDeployment      = "Deployment"
		myCRD             = "myCRD"
		expectedResultSMP = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - env:
        - name: SOMEENV
          value: SOMEVALUE
        image: nginx
        name: nginx
`
	)
	tests := map[string]struct {
		base          []string
		patches       []string
		expected      []string
		errorExpected bool
		errorMsg      string
	}{
		"clown": {
			base: []string{`apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`,
			},
			patches: []string{`apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 999
`,
			},
			errorExpected: false,
			expected: []string{
				`apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 999
`,
			},
		},
		"confusion": {
			base: []string{`apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    A: X
    B: Y
`,
			},
			patches: []string{`apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
`, `apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: Z
    D: W
  baz:
    hello: world
`,
			},
			errorExpected: false,
			expected: []string{
				`apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    A: X
    C: Z
    D: W
  baz:
    hello: world
`,
			},
		},
		"withschema-ns1-ns2-one": {
			base: []string{
				addNamespace("ns1", baseResource(myDeployment)),
				addNamespace("ns2", baseResource(myDeployment)),
			},
			patches: []string{
				addNamespace("ns1", addLabelAndEnvPatch(myDeployment)),
				addNamespace("ns2", addLabelAndEnvPatch(myDeployment)),
			},
			errorExpected: false,
			expected: []string{
				addNamespace("ns1", expectedResultSMP),
				addNamespace("ns2", expectedResultSMP),
			},
		},
		"withschema-ns1-ns2-two": {
			base: []string{
				addNamespace("ns1", baseResource(myDeployment)),
			},
			patches: []string{
				addNamespace("ns2", changeImagePatch(myDeployment)),
			},
			expected: []string{
				addNamespace("ns1", baseResource(myDeployment)),
			},
		},
		"withschema-ns1-ns2-three": {
			base: []string{
				addNamespace("ns1", baseResource(myDeployment)),
			},
			patches: []string{
				addNamespace("ns1", changeImagePatch(myDeployment)),
			},
			expected: []string{
				`apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
  namespace: ns1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
`,
			},
		},
		"withschema-nil-ns2": {
			base: []string{
				baseResource(myDeployment),
			},
			patches: []string{
				addNamespace("ns2", changeImagePatch(myDeployment)),
			},
			expected: []string{
				baseResource(myDeployment),
			},
		},
		"withschema-ns1-nil": {
			base: []string{
				addNamespace("ns1", baseResource(myDeployment)),
			},
			patches: []string{
				changeImagePatch(myDeployment),
			},
			expected: []string{
				addNamespace("ns1", baseResource(myDeployment)),
			},
		},
		"noschema-ns1-ns2-one": {
			base: []string{
				addNamespace("ns1", baseResource(myCRD)),
				addNamespace("ns2", baseResource(myCRD)),
			},
			patches: []string{
				addNamespace("ns1", addLabelAndEnvPatch(myCRD)),
				addNamespace("ns2", addLabelAndEnvPatch(myCRD)),
			},
			errorExpected: false,
			expected: []string{
				addNamespace("ns1", expectedResultJMP("")),
				addNamespace("ns2", expectedResultJMP("")),
			},
		},
		"noschema-ns1-ns2-two": {
			base:     []string{addNamespace("ns1", baseResource(myCRD))},
			patches:  []string{addNamespace("ns2", changeImagePatch(myCRD))},
			expected: []string{addNamespace("ns1", baseResource(myCRD))},
		},
		"noschema-nil-ns2": {
			base:     []string{baseResource(myCRD)},
			patches:  []string{addNamespace("ns2", changeImagePatch(myCRD))},
			expected: []string{baseResource(myCRD)},
		},
		"noschema-ns1-nil": {
			base:     []string{addNamespace("ns1", baseResource(myCRD))},
			patches:  []string{changeImagePatch(myCRD)},
			expected: []string{addNamespace("ns1", baseResource(myCRD))},
		},
	}
	for n := range tests {
		tc := tests[n]
		t.Run(n, func(t *testing.T) {
			m, err := rmF.NewResMapFromBytes([]byte(strings.Join(tc.base, "\n---\n")))
			assert.NoError(t, err)
			foundError := false
			for _, patch := range tc.patches {
				rp, err := rf.FromBytes([]byte(patch))
				assert.NoError(t, err)
				idSet := resource.MakeIdSet([]*resource.Resource{rp})
				if err = m.ApplySmPatch(idSet, rp); err != nil {
					foundError = true
					break
				}
			}
			if foundError {
				assert.True(t, tc.errorExpected)
				// compare error message?
				return
			}
			assert.False(t, tc.errorExpected)
			m.RemoveBuildAnnotations()
			yml, err := m.AsYaml()
			assert.NoError(t, err)
			assert.Equal(t, strings.Join(tc.expected, "---\n"), string(yml))
		})
	}
}

// simple utility function to add an namespace in a resource
// used as base, patch or expected result. Simply looks
// for specs: in order to add namespace: xxxx before this line
func addNamespace(namespace string, base string) string {
	res := strings.Replace(base,
		"\nspec:\n",
		"\n  namespace: "+namespace+"\nspec:\n",
		1)
	return res
}

// DeleteOddsFilter deletes the odd entries, removing nodes.
// This is a ridiculous filter for testing.
type DeleteOddsFilter struct{}

func (f DeleteOddsFilter) Filter(
	nodes []*yaml.RNode) (result []*yaml.RNode, err error) {
	for i := range nodes {
		if i%2 == 0 {
			// Keep the even entries, drop the odd entries.
			result = append(result, nodes[i])
		}
	}
	return
}

// CloneOddsFilter deletes even entries and clones odd entries,
// making new nodes.
// This is a ridiculous filter for testing.
type CloneOddsFilter struct{}

func (f CloneOddsFilter) Filter(
	nodes []*yaml.RNode) (result []*yaml.RNode, err error) {
	for i := range nodes {
		if i%2 != 0 {
			newNode := nodes[i].Copy()
			// Add suffix to the name, so that it's unique (w/r to this test).
			newNode.SetName(newNode.GetName() + "Clone")
			// Return a ptr to the copy.
			result = append(result, nodes[i], newNode)
		}
	}
	return
}

func TestApplyFilter(t *testing.T) {
	tests := map[string]struct {
		input    string
		f        kio.Filter
		expected string
	}{
		"labels": {
			input: `
apiVersion: example.com/v1
kind: Beans
metadata:
  name: myBeans
---
apiVersion: example.com/v1
kind: Franks
metadata:
  name: myFranks
`,
			f: labels.Filter{
				Labels: map[string]string{
					"a": "foo",
					"b": "bar",
				},
				FsSlice: types.FsSlice{
					{
						Gvk:                resid.NewGvk("example.com", "v1", "Beans"),
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
				},
			},
			expected: `
apiVersion: example.com/v1
kind: Beans
metadata:
  labels:
    a: foo
    b: bar
  name: myBeans
---
apiVersion: example.com/v1
kind: Franks
metadata:
  name: myFranks
`,
		},
		"deleteOddNodes": {
			input: `
apiVersion: example.com/v1
kind: Zero
metadata:
  name: r0
---
apiVersion: example.com/v1
kind: One
metadata:
  name: r1
---
apiVersion: example.com/v1
kind: Two
metadata:
  name: r2
---
apiVersion: example.com/v1
kind: Three
metadata:
  name: r3
`,
			f: DeleteOddsFilter{},
			expected: `
apiVersion: example.com/v1
kind: Zero
metadata:
  name: r0
---
apiVersion: example.com/v1
kind: Two
metadata:
  name: r2
`,
		},
		"cloneOddNodes": {
			// input list has five entries
			input: `
apiVersion: example.com/v1
kind: Zero
metadata:
  name: r0
---
apiVersion: example.com/v1
kind: One
metadata:
  name: r1
---
apiVersion: example.com/v1
kind: Two
metadata:
  name: r2
---
apiVersion: example.com/v1
kind: Three
metadata:
  name: r3
---
apiVersion: example.com/v1
kind: Four
metadata:
  name: r4
`,
			f: CloneOddsFilter{},
			// output has four, but half are newly created nodes.
			expected: `
apiVersion: example.com/v1
kind: One
metadata:
  name: r1
---
apiVersion: example.com/v1
kind: One
metadata:
  name: r1Clone
---
apiVersion: example.com/v1
kind: Three
metadata:
  name: r3
---
apiVersion: example.com/v1
kind: Three
metadata:
  name: r3Clone
`,
		},
	}
	for name := range tests {
		tc := tests[name]
		t.Run(name, func(t *testing.T) {
			m, err := rmF.NewResMapFromBytes([]byte(tc.input))
			assert.NoError(t, err)
			assert.NoError(t, m.ApplyFilter(tc.f))
			kusttest_test.AssertActualEqualsExpectedWithTweak(
				t, m, nil, tc.expected)
		})
	}
}

func TestApplySmPatch_Deletion(t *testing.T) {
	target := `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
	tests := map[string]struct {
		patch        string
		expected     string
		finalMapSize int
	}{
		"delete1": {
			patch: `apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 2
  template:
    $patch: delete
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 2
`,
			finalMapSize: 1,
		},
		"delete2": {
			patch: `apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  $patch: delete
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
`,
			finalMapSize: 1,
		},
		"delete3": {
			patch: `apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
$patch: delete
`,
			expected:     "",
			finalMapSize: 0,
		},
	}
	for name := range tests {
		tc := tests[name]
		t.Run(name, func(t *testing.T) {
			m, err := rmF.NewResMapFromBytes([]byte(target))
			assert.NoError(t, err, name)
			idSet := resource.MakeIdSet(m.Resources())
			assert.Equal(t, 1, idSet.Size(), name)
			p, err := rf.FromBytes([]byte(tc.patch))
			assert.NoError(t, err, name)
			assert.NoError(t, m.ApplySmPatch(idSet, p), name)
			assert.Equal(t, tc.finalMapSize, m.Size(), name)
			m.RemoveBuildAnnotations()
			yml, err := m.AsYaml()
			assert.NoError(t, err, name)
			assert.Equal(t, tc.expected, string(yml), name)
		})
	}
}

// baseResource produces a base object which used to test
// patch transformation
// Also the structure is matching the Deployment syntax
// the kind can be replaced to allow testing using CRD
// without access to the schema
func baseResource(kind string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`, kind)
}

// addContainerAndEnvPatch produces a patch object which adds
// an entry in the env slice of the first/nginx container
// as well as adding a label in the metadata
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func addLabelAndEnvPatch(kind string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        some-label: some-value
    spec:
      containers:
       - name: nginx
         env:
         - name: SOMEENV
           value: SOMEVALUE`, kind)
}

// changeImagePatch produces a patch object which replaces
// the value of the image field in the first/nginx container
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func changeImagePatch(kind string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: "nginx:1.7.9"`, kind)
}

// utility method building the expected output of a JMP.
// imagename parameter allows to build a result consistent
// with the JMP behavior which basically overrides the
// entire "containers" list.
func expectedResultJMP(imagename string) string {
	if imagename == "" {
		return `apiVersion: apps/v1
kind: myCRD
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - env:
        - name: SOMEENV
          value: SOMEVALUE
        name: nginx
`
	}
	return fmt.Sprintf(`apiVersion: apps/v1
kind: myCRD
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - image: %s
        name: nginx
`, imagename)
}
