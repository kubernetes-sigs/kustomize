// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resid

import (
	"testing"
)

var resIdStringTests = []struct {
	x ResId
	s string
}{
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"g_v_k|ns|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"~G_v_k|ns|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{Kind: "k"},
			Name:      "nm",
		},
		"~G_~V_k|ns|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{},
			Name:      "nm",
		},
		"~G_~V_~K|ns|nm",
	},
	{
		ResId{
			Gvk:  Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|~X|nm",
	},
	{
		ResId{
			Gvk:  Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|~X|nm",
	},
	{
		ResId{
			Gvk: Gvk{},
		},
		"~G_~V_~K|~X|~N",
	},
	{
		ResId{
			Gvk: Gvk{},
		},
		"~G_~V_~K|~X|~N",
	},
	{
		ResId{},
		"~G_~V_~K|~X|~N",
	},
}

func TestResIdString(t *testing.T) {
	for _, hey := range resIdStringTests {
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
			Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"g_v_k|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{Version: "v", Kind: "k"},
			Name:      "nm",
		},
		"~G_v_k|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{Kind: "k"},
			Name:      "nm",
		},
		"~G_~V_k|nm",
	},
	{
		ResId{
			Namespace: "ns",
			Gvk:       Gvk{},
			Name:      "nm",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			Gvk:  Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			Gvk:  Gvk{},
			Name: "nm",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			Gvk: Gvk{},
		},
		"~G_~V_~K|",
	},
	{
		ResId{
			Gvk: Gvk{},
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

func TestResIdEquals(t *testing.T) {

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
				Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "X",
				Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   true,
			equals:     true,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Gvk:  Gvk{Group: "g", Version: "v", Kind: "k"},
				Name: "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Gvk:       Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "X",
				Gvk:       Gvk{Kind: "k"},
				Name:      "nm",
			},
			id2: ResId{
				Namespace: "Z",
				Gvk:       Gvk{Kind: "k"},
				Name:      "nm",
			},
			gVknResult: true,
			nsEquals:   false,
			equals:     false,
		},
		{
			id1: ResId{
				Gvk:  Gvk{Kind: "k"},
				Name: "nm",
			},
			id2: ResId{
				Gvk:  Gvk{Kind: "k"},
				Name: "nm2",
			},
			gVknResult: false,
			nsEquals:   true,
			equals:     false,
		},
		{
			id1: ResId{
				Gvk:  Gvk{Kind: "k"},
				Name: "nm",
			},
			id2: ResId{
				Gvk:  Gvk{Kind: "Node"},
				Name: "nm",
			},
			gVknResult: false,
			nsEquals:   true,
			equals:     false,
		},
		{
			id1: ResId{
				Gvk:  Gvk{Kind: "Node"},
				Name: "nm1",
			},
			id2: ResId{
				Gvk:  Gvk{Kind: "Node"},
				Name: "nm2",
			},
			gVknResult: false,
			nsEquals:   true,
			equals:     false,
		},
		{
			id1: ResId{
				Namespace: "default",
				Gvk:       Gvk{Kind: "k"},
				Name:      "nm1",
			},
			id2: ResId{
				Gvk:  Gvk{Kind: "k"},
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
		Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       Gvk{Version: "v", Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       Gvk{Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       Gvk{},
		Name:      "nm",
	},
	{
		Gvk:  Gvk{},
		Name: "nm",
	},
	{
		Gvk:  Gvk{},
		Name: "nm",
	},
	{
		Gvk: Gvk{},
	},
}

func TestResIdIsSelected(t *testing.T) {
	type selectable struct {
		id             ResId
		expectSelected bool
	}
	var testCases = []struct {
		selector    ResId
		selectables []selectable
	}{
		{
			selector: ResId{
				Namespace: "X",
				Name:      "nm",
				Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
			},
			selectables: []selectable{
				{
					id: ResId{
						Namespace: "X",
						Name:      "nm",
						Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
					},
					expectSelected: true,
				},
				{
					id: ResId{
						Namespace: "x",
						Name:      "nm",
						Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
					},
					expectSelected: false,
				},
				{
					id: ResId{
						Name: "nm",
						Gvk:  Gvk{Group: "g", Version: "v", Kind: "k"},
					},
					expectSelected: false,
				},
			},
		},
		{
			selector: ResId{
				/* Namespace wildcard */
				Name: "nm",
				Gvk:  Gvk{Group: "g" /* Version wildcard */, Kind: "k"},
			},
			selectables: []selectable{
				{
					id: ResId{
						Namespace: "X",
						Name:      "nm",
						Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
					},
					expectSelected: true,
				},
				{
					id: ResId{
						Namespace: "x",
						Name:      "nm",
						Gvk:       Gvk{Group: "g", Version: "v", Kind: "k"},
					},
					expectSelected: true,
				},
				{
					id: ResId{
						Name: "nm",
						Gvk:  Gvk{Group: "g", Version: "VVV", Kind: "k"},
					},
					expectSelected: true,
				},
			},
		},
	}

	for _, tst := range testCases {
		for _, pair := range tst.selectables {
			if pair.id.IsSelectedBy(tst.selector) {
				if !pair.expectSelected {
					t.Fatalf(
						"expected id %s to NOT be selected by %s",
						pair.id, tst.selector)
				}
			} else {
				if pair.expectSelected {
					t.Fatalf(
						"expected id %s to be selected by %s",
						pair.id, tst.selector)
				}
			}
		}
	}
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
	var testCases = map[string]struct {
		id       ResId
		expected string
	}{
		"tst1": {
			id: ResId{
				Gvk:  NewGvk("", "v1", "Node"),
				Name: "nm",
			},
			expected: TotallyNotANamespace,
		},
		"tst2": {
			id: ResId{
				Namespace: "foo",
				Gvk:       NewGvk("", "v1", "Node"),
				Name:      "nm",
			},
			expected: TotallyNotANamespace,
		},
		"tst3": {
			id: ResId{
				Namespace: "foo",
				Gvk:       NewGvk("g", "v", "k"),
				Name:      "nm",
			},
			expected: "foo",
		},
		"tst4": {
			id: ResId{
				Namespace: "",
				Gvk:       NewGvk("g", "v", "k"),
				Name:      "nm",
			},
			expected: DefaultNamespace,
		},
		"tst5": {
			id: ResId{
				Gvk:  Gvk{Group: "g", Version: "v", Kind: "k"},
				Name: "nm",
			},
			expected: DefaultNamespace,
		},
	}

	for n, tst := range testCases {
		t.Run(n, func(t *testing.T) {
			if actual := tst.id.EffectiveNamespace(); actual != tst.expected {
				t.Fatalf("EffectiveNamespace was %s, expected %s",
					actual, tst.expected)
			}
		})
	}
}
