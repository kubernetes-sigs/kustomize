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
		"noGroup_v_k|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"noGroup_noVersion_k|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"noGroup_noVersion_noKind|ns|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s"},
		"noGroup_noVersion_noKind|noNamespace|p|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", suffix: "s"},
		"noGroup_noVersion_noKind|noNamespace|noPrefix|nm|s"},
	{ResId{gvKind: gvk.Gvk{},
		suffix: "s"},
		"noGroup_noVersion_noKind|noNamespace|noPrefix|noName|s"},
	{ResId{gvKind: gvk.Gvk{}},
		"noGroup_noVersion_noKind|noNamespace|noPrefix|noName|noSuffix"},
	{ResId{},
		"noGroup_noVersion_noKind|noNamespace|noPrefix|noName|noSuffix"},
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
		"noGroup_v_k|nm"},
	{ResId{gvKind: gvk.Gvk{Kind: "k"},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"noGroup_noVersion_k|nm"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s", namespace: "ns"},
		"noGroup_noVersion_noKind|nm"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", prefix: "p", suffix: "s"},
		"noGroup_noVersion_noKind|nm"},
	{ResId{gvKind: gvk.Gvk{},
		name: "nm", suffix: "s"},
		"noGroup_noVersion_noKind|nm"},
	{ResId{gvKind: gvk.Gvk{},
		suffix: "s"},
		"noGroup_noVersion_noKind|"},
	{ResId{gvKind: gvk.Gvk{}},
		"noGroup_noVersion_noKind|"},
	{ResId{},
		"noGroup_noVersion_noKind|"},
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
