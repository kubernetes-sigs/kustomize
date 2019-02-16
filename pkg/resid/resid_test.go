package resid

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

var stringTests = []struct {
	x ResId
	s string
}{
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"g_v_k|ns|p|nm|s",
	},
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"~G_v_k|ns|p|nm|s",
	},
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{Kind: "k"},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"~G_~V_k|ns|p|nm|s",
	},
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"~G_~V_~K|ns|p|nm|s",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
			name:   "nm",
			prefix: "p",
			suffix: "s",
		},
		"~G_~V_~K|~X|p|nm|s",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
			name:   "nm",
			suffix: "s",
		},
		"~G_~V_~K|~X|~P|nm|s",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
			suffix: "s",
		},
		"~G_~V_~K|~X|~P|~N|s",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
		},
		"~G_~V_~K|~X|~P|~N|~S",
	},
	{
		ResId{},
		"~G_~V_~K|~X|~P|~N|~S",
	},
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
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"g_v_k|nm",
	},
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"~G_v_k|nm",
	},
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{Kind: "k"},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"~G_~V_k|nm",
	},
	{
		ResId{
			namespace: "ns",
			gvKind:    gvk.Gvk{},
			name:      "nm",
			prefix:    "p",
			suffix:    "s",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
			name:   "nm",
			prefix: "p",
			suffix: "s",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
			name:   "nm",
			suffix: "s",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
			suffix: "s",
		},
		"~G_~V_~K|",
	},
	{
		ResId{
			gvKind: gvk.Gvk{},
		},
		"~G_~V_~K|",
	},
	{
		ResId{},
		"~G_~V_~K|",
	},
}

func TestGvknString(t *testing.T) {
	for _, hey := range gvknStringTests {
		if hey.x.GvknString() != hey.s {
			t.Fatalf("Actual: %s,  Expected: '%s'", hey.x.GvknString(), hey.s)
		}
	}
}

var GvknEqualsTest = []struct {
	id1          ResId
	id2          ResId
	gVknResult   bool
	nSgVknResult bool
}{
	{
		id1: ResId{
			namespace: "X",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "AA",
			suffix:    "aa",
		},
		id2: ResId{
			namespace: "X",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "BB",
			suffix:    "bb",
		},
		gVknResult:   true,
		nSgVknResult: true,
	},
	{
		id1: ResId{
			namespace: "X",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "AA",
			suffix:    "aa",
		},
		id2: ResId{
			namespace: "Z",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "BB",
			suffix:    "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			namespace: "X",
			gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "AA",
			suffix:    "aa",
		},
		id2: ResId{
			gvKind: gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			name:   "nm",
			prefix: "BB",
			suffix: "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			namespace: "X",
			gvKind:    gvk.Gvk{Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "AA",
			suffix:    "aa",
		},
		id2: ResId{
			namespace: "Z",
			gvKind:    gvk.Gvk{Version: "v", Kind: "k"},
			name:      "nm",
			prefix:    "BB",
			suffix:    "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			namespace: "X",
			gvKind:    gvk.Gvk{Kind: "k"},
			name:      "nm",
			prefix:    "AA",
			suffix:    "aa",
		},
		id2: ResId{
			namespace: "Z",
			gvKind:    gvk.Gvk{Kind: "k"},
			name:      "nm",
			prefix:    "BB",
			suffix:    "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			namespace: "X",
			name:      "nm",
			prefix:    "AA",
			suffix:    "aa",
		},
		id2: ResId{
			namespace: "Z",
			name:      "nm",
			prefix:    "BB",
			suffix:    "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
}

func TestEquals(t *testing.T) {
	for _, tst := range GvknEqualsTest {
		if tst.id1.GvknEquals(tst.id2) != tst.gVknResult {
			t.Fatalf("GvknEquals(\n%v,\n%v\n) should be %v",
				tst.id1, tst.id2, tst.gVknResult)
		}
		if tst.id1.NsGvknEquals(tst.id2) != tst.nSgVknResult {
			t.Fatalf("NsGvknEquals(\n%v,\n%v\n) should be %v",
				tst.id1, tst.id2, tst.nSgVknResult)
		}
	}
}

func TestCopyWithNewPrefixSuffix(t *testing.T) {
	r1 := ResId{
		namespace: "X",
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "a",
		suffix:    "b",
	}
	r2 := r1.CopyWithNewPrefixSuffix("p-", "-s")
	expected := ResId{
		namespace: "X",
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "p-a",
		suffix:    "b-s",
	}
	if !r2.GvknEquals(expected) {
		t.Fatalf("%v should equal %v", r2, expected)
	}
}

func TestCopyWithNewNamespace(t *testing.T) {
	r1 := ResId{
		namespace: "X",
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "a",
		suffix:    "b",
	}
	r2 := r1.CopyWithNewNamespace("zzz")
	expected := ResId{
		namespace: "zzz",
		gvKind:    gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		name:      "nm",
		prefix:    "a",
		suffix:    "b",
	}
	if !r2.GvknEquals(expected) {
		t.Fatalf("%v should equal %v", r2, expected)
	}
}
