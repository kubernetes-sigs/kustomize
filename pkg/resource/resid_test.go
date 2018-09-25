package resource

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

var stringTests = []struct {
	x ResId
	s string
}{
	{ResId{gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name: "nm", prefix: "p", namespace: "ns"}, "g_v_k_ns_p_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
		name: "nm", prefix: "p", namespace: "ns"}, "_v_k_ns_p_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "p", namespace: "ns"}, "__k_ns_p_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", namespace: "ns"}, "___ns_p_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p"}, "____p_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm"}, "_____nm.yaml"},
	{ResId{gvKind: gvk.Gvk{}}, "_____.yaml"},
	{ResId{}, "_____.yaml"},
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
		name: "nm", prefix: "p", namespace: "ns"}, "g_v_k_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
		name: "nm", prefix: "p", namespace: "ns"}, "v_k_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "p", namespace: "ns"}, "_k_nm.yaml"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", namespace: "ns"}, "__nm.yaml"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p"}, "__nm.yaml"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm"}, "__nm.yaml"},
	{ResId{gvKind: gvk.Gvk{}}, "__.yaml"},
	{ResId{}, "__.yaml"},
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
		name: "nm", prefix: "AA", namespace: "X"},
		ResId{gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name: "nm", prefix: "BB", namespace: "Z"}},
	{ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
		name: "nm", prefix: "AA", namespace: "X"},
		ResId{gvKind: gvk.Gvk{Version: "v", Kind: "k"},
			name: "nm", prefix: "BB", namespace: "Z"}},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "AA", namespace: "X"},
		ResId{gvKind: gvk.Gvk{Kind: "k"},
			name: "nm", prefix: "BB", namespace: "Z"}},
	{ResId{name: "nm", prefix: "AA", namespace: "X"},
		ResId{name: "nm", prefix: "BB", namespace: "Z"}},
}

func TestEquals(t *testing.T) {
	for _, hey := range GvknEqualsTest {
		if !hey.x1.GvknEquals(hey.x2) {
			t.Fatalf("%v should equal %v", hey.x1, hey.x2)
		}
	}
}
