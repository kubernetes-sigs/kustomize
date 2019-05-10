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
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"g_v_k|ns|p|nm|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_v_k|ns|p|nm|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_~V_k|ns|p|nm|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_~V_~K|ns|p|nm|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk:  gvk.Gvk{},
				Name: "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_~V_~K|~X|p|nm|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk:  gvk.Gvk{},
				Name: "nm",
			},
			Suffix: "s",
		},
		"~G_~V_~K|~X|~P|nm|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk: gvk.Gvk{},
			},
			Suffix: "s",
		},
		"~G_~V_~K|~X|~P|~N|s",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk: gvk.Gvk{},
			},
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
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"g_v_k|nm",
	},
	{
		ResId{
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_v_k|nm",
	},
	{
		ResId{
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_~V_k|nm",
	},
	{
		ResId{
			ItemId: ItemId{
				Namespace: "ns",
				Gvk:       gvk.Gvk{},
				Name:      "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk:  gvk.Gvk{},
				Name: "nm",
			},
			Prefix: "p",
			Suffix: "s",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk:  gvk.Gvk{},
				Name: "nm",
			},
			Suffix: "s",
		},
		"~G_~V_~K|nm",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk: gvk.Gvk{},
			},
			Suffix: "s",
		},
		"~G_~V_~K|",
	},
	{
		ResId{
			ItemId: ItemId{
				Gvk: gvk.Gvk{},
			},
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
			ItemId: ItemId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "AA",
			Suffix: "aa",
		},
		id2: ResId{
			ItemId: ItemId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "BB",
			Suffix: "bb",
		},
		gVknResult:   true,
		nSgVknResult: true,
	},
	{
		id1: ResId{
			ItemId: ItemId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "AA",
			Suffix: "aa",
		},
		id2: ResId{
			ItemId: ItemId{
				Namespace: "Z",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "BB",
			Suffix: "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			ItemId: ItemId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "AA",
			Suffix: "aa",
		},
		id2: ResId{
			ItemId: ItemId{
				Gvk:  gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
				Name: "nm",
			},
			Prefix: "BB",
			Suffix: "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			ItemId: ItemId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "AA",
			Suffix: "aa",
		},
		id2: ResId{
			ItemId: ItemId{
				Namespace: "Z",
				Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
				Name:      "nm",
			},
			Prefix: "BB",
			Suffix: "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			ItemId: ItemId{
				Namespace: "X",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm",
			},
			Prefix: "AA",
			Suffix: "aa",
		},
		id2: ResId{
			ItemId: ItemId{
				Namespace: "Z",
				Gvk:       gvk.Gvk{Kind: "k"},
				Name:      "nm",
			},
			Prefix: "BB",
			Suffix: "bb",
		},
		gVknResult:   true,
		nSgVknResult: false,
	},
	{
		id1: ResId{
			ItemId: ItemId{
				Namespace: "X",
				Name:      "nm",
			},
			Prefix: "AA",
			Suffix: "aa",
		},
		id2: ResId{
			ItemId: ItemId{
				Namespace: "Z",
				Name:      "nm",
			},
			Prefix: "BB",
			Suffix: "bb",
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
		ItemId: ItemId{
			Namespace: "X",
			Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		Prefix: "a",
		Suffix: "b",
	}
	r2 := r1.CopyWithNewPrefixSuffix("p-", "-s")
	expected := ResId{
		ItemId: ItemId{
			Namespace: "X",
			Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		Prefix: "p-a",
		Suffix: "b-s",
	}
	if !r2.GvknEquals(expected) {
		t.Fatalf("%v should equal %v", r2, expected)
	}
}

func TestCopyWithNewNamespace(t *testing.T) {
	r1 := ResId{
		ItemId: ItemId{
			Namespace: "X",
			Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		Prefix: "a",
		Suffix: "b",
	}
	r2 := r1.CopyWithNewNamespace("zzz")
	expected := ResId{
		ItemId: ItemId{
			Namespace: "zzz",
			Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
			Name:      "nm",
		},
		Prefix: "a",
		Suffix: "b",
	}
	if !r2.GvknEquals(expected) {
		t.Fatalf("%v should equal %v", r2, expected)
	}
}
