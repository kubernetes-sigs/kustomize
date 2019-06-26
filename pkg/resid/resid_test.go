// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resid

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
)

var stringTests = []struct {
	x ResId
	s string
}{
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"g_v_k|ns|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"~G_v_k|ns|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{Kind: "k"},
			Name:      "nm",
		},
		"~G_~V_k|ns|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{},
			Name:      "nm",
		},
		"~G_~V_~K|ns|nm",
	},
	{
		ResId{
			Gvk:  gvk.Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|~X|nm",
	},
	{
		ResId{
			Gvk:  gvk.Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|~X|nm",
	},
	{
		ResId{
			Gvk: gvk.Gvk{},
		},
		"~G_~V_~K|~X|~N",
	},
	{
		ResId{
			Gvk: gvk.Gvk{},
		},
		"~G_~V_~K|~X|~N",
	},
	{
		ResId{},
		"~G_~V_~K|~X|~N",
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
			Namespace: "ns",
			Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"g_v_k|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"~G_v_k|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{Kind: "k"},
			Name:      "nm",
		},
		"~G_~V_k|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       gvk.Gvk{},
			Name:      "nm",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			Gvk:  gvk.Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			Gvk:  gvk.Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			Gvk: gvk.Gvk{},
		},
		"~G_~V_~K|",
	},
	{
		ResId{
			Gvk: gvk.Gvk{},
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

func TestEquals(t *testing.T) {

	var GvknEqualsTest = []struct {
		id1        ResId
		id2        ResId
		gVknResult bool
		nsEquals   bool
		equals     bool
	}{
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   true,
			equals:     true,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Gvk:  gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name: "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Gvk:  gvk.Gvk{Kind: "k"},
				Name: "nm",
			},
			id2: ResId{
				Gvk:  gvk.Gvk{Kind: "k"},
				Name: "nm2",
			},
			gVknResult: false,
			nsEquals:   true,
			equals:     false,
		},
		{
			id1: ResId{
				Gvk:  gvk.Gvk{Kind: "k"},
				Name: "nm",
			},
			id2: ResId{
				Gvk:  gvk.Gvk{Kind: "Node"},
				Name: "nm",
			},
			gVknResult: false,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Gvk:  gvk.Gvk{Kind: "Node"},
				Name: "nm1",
			},
			id2: ResId{
				Gvk:  gvk.Gvk{Kind: "Node"},
				Name: "nm2",
			},
			gVknResult: false,
			nsEquals:   true,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "default",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm1",
			},
			id2: ResId{
				Gvk:  gvk.Gvk{Kind: "k"},
				Name: "nm2",
			},
			gVknResult: false,
			nsEquals:   true,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
	}

	for _, tst := range GvknEqualsTest {
		if tst.id1.GvknEquals(tst.id2) != tst.gVknResult {
			t.Fatalf("GvknEquals(\n%v,\n%v\n) should be %v",
				tst.id1, tst.id2, tst.gVknResult)
		}
		if tst.id1.IsNsEquals(tst.id2) != tst.nsEquals {
			t.Fatalf("IsNsEquals(\n%v,\n%v\n) should be %v",
				tst.id1, tst.id2, tst.equals)
		}
		if tst.id1.Equals(tst.id2) != tst.equals {
			t.Fatalf("Equals(\n%v,\n%v\n) should be %v",
				tst.id1, tst.id2, tst.equals)
		}
	}
}

var ids = []ResId{
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{},
		Name:      "nm",
	},
	{
		Gvk:  gvk.Gvk{},
		Name: "nm",
	},
	{
		Gvk:  gvk.Gvk{},
		Name: "nm",
	},
	{
		Gvk: gvk.Gvk{},
	},
}

func TestFromString(t *testing.T) {
	for _, id := range ids {
		newId := FromString(id.String())
		if newId != id {
			t.Fatalf("Actual: %v,  Expected: '%s'", newId, id)
		}
	}
}

func TestEffectiveNamespace(t *testing.T) {
	var test = []struct {
		id       ResId
		expected string
	}{
		{
			id: ResId{
				Gvk:  gvk.Gvk{Group: "g", Version: "v", Kind: "Node"},
				Name: "nm",
			},
			expected: TotallyNotANamespace,
		},
		{
			id: ResId{
				Namespace: "foo",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "Node"},
				Name:      "nm",
			},
			expected: TotallyNotANamespace,
		},
		{
			id: ResId{
				Namespace: "foo",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			expected: "foo",
		},
		{
			id: ResId{
				Namespace: "",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			expected: DefaultNamespace,
		},
		{
			id: ResId{
				Gvk:  gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name: "nm",
			},
			expected: DefaultNamespace,
		},
	}

	for _, tst := range test {
		if actual := tst.id.EffectiveNamespace(); actual != tst.expected {
			t.Fatalf("EffectiveNamespace was %s, expected %s",
				actual, tst.expected)
		}
	}
}
