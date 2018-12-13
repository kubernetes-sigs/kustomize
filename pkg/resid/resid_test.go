package resid

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

var stringTests = []struct {
	x ResId
	s string
}{
	{ResId{gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"g_v_k|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"~G_v_k|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"~G_~V_k|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"~G_~V_~K|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s"},
		"~G_~V_~K|~X|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", suffix: "s"},
		"~G_~V_~K|~X|~P|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		suffix: "s"},
		"~G_~V_~K|~X|~P|~N|s"},
	{ResId{gvKind: gvk.Gvk{}},
		"~G_~V_~K|~X|~P|~N|~S"},
	{ResId{},
		"~G_~V_~K|~X|~P|~N|~S"},
}

func TestString(t *testing.T) {
	for _, hey := range stringTests {
		if hey.x.String() != hey.s {
			t.Fatalf("Actual: %v,  Expected: '%s'", hey.x, hey.s)
		}
	}
}

var gvknStringTests = []struct {
	x ResId
	s string
}{
	{ResId{gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"g_v_k|nm"},
	{ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"~G_v_k|nm"},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"~G_~V_k|nm"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"~G_~V_~K|nm"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s"},
		"~G_~V_~K|nm"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", suffix: "s"},
		"~G_~V_~K|nm"},
	{ResId{gvKind: gvk.Gvk{},
		suffix: "s"},
		"~G_~V_~K|"},
	{ResId{gvKind: gvk.Gvk{}},
		"~G_~V_~K|"},
	{ResId{},
		"~G_~V_~K|"},
}

func TestGvknString(t *testing.T) {
	for _, hey := range gvknStringTests {
		if hey.x.GvknString() != hey.s {
			t.Fatalf("Actual: %s,  Expected: '%s'", hey.x.GvknString(), hey.s)
		}
	}
}

var GvknEqualsTest = []struct {
	x1 ResId
	x2 ResId
}{
	{ResId{gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name: "nm", prefix: "AA", suffix: "aa", namespace: "X"},
		ResId{gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name: "nm", prefix: "BB", suffix: "bb", namespace: "Z"}},
	{ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
		name: "nm", prefix: "AA", suffix: "aa", namespace: "X"},
		ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
			name: "nm", prefix: "BB", suffix: "bb", namespace: "Z"}},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "AA", suffix: "aa", namespace: "X"},
		ResId{gvKind: gvk.Gvk{Kind: "k"},
			name: "nm", prefix: "BB", suffix: "bb", namespace: "Z"}},
	{ResId{name: "nm", prefix: "AA", suffix: "aa", namespace: "X"},
		ResId{name: "nm", prefix: "BB", suffix: "bb", namespace: "Z"}},
}

func TestEquals(t *testing.T) {
	for _, hey := range GvknEqualsTest {
		if !hey.x1.GvknEquals(hey.x2) {
			t.Fatalf("%v should equal %v", hey.x1, hey.x2)
		}
	}
}

func TestCopyWithNewPrefixSuffix(t *testing.T) {
	r1 := ResId{
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "a",
		suffix:    "b",
		namespace: "X"}
	r2 := r1.CopyWithNewPrefixSuffix("p-", "-s")
	expected := ResId{
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "p-a",
		suffix:    "b-s",
		namespace: "X"}
	if !r2.GvknEquals(expected) {
		t.Fatalf("%v should equal %v", r2, expected)
	}
}

func TestCopyWithNewNamespace(t *testing.T) {
	r1 := ResId{
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "a",
		suffix:    "b",
		namespace: "X"}
	r2 := r1.CopyWithNewNamespace("zzz")
	expected := ResId{
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "a",
		suffix:    "b",
		namespace: "zzz"}
	if !r2.GvknEquals(expected) {
		t.Fatalf("%v should equal %v", r2, expected)
	}
}
