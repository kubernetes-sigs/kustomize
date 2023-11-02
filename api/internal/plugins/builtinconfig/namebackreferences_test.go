// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinconfig

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func TestMergeAll(t *testing.T) {
	fsSlice1 := []types.FieldSpec{
		{
			Gvk: resid.Gvk{
				Kind: "Pod",
			},
			Path:               "path/to/a/name",
			CreateIfNotPresent: false,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Deployment",
			},
			Path:               "another/path/to/some/name",
			CreateIfNotPresent: false,
		},
	}
	fsSlice2 := []types.FieldSpec{
		{
			Gvk: resid.Gvk{
				Kind: "Job",
			},
			Path:               "morepath/to/name",
			CreateIfNotPresent: false,
		},
		{
			Gvk: resid.Gvk{
				Kind: "StatefulSet",
			},
			Path:               "yet/another/path/to/a/name",
			CreateIfNotPresent: false,
		},
	}

	nbrsSlice1 := nbrSlice{
		{
			Gvk: resid.Gvk{
				Kind: "ConfigMap",
			},
			Referrers: fsSlice1,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Secret",
			},
			Referrers: fsSlice2,
		},
	}
	nbrsSlice2 := nbrSlice{
		{
			Gvk: resid.Gvk{
				Kind: "ConfigMap",
			},
			Referrers: fsSlice1,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Secret",
			},
			Referrers: fsSlice2,
		},
	}
	expected := nbrSlice{
		{
			Gvk: resid.Gvk{
				Kind: "ConfigMap",
			},
			Referrers: fsSlice1,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Secret",
			},
			Referrers: fsSlice2,
		},
	}
	actual, err := nbrsSlice1.mergeAll(nbrsSlice2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected\n %v\n but got\n %v\n", expected, actual)
	}
}

func TestNbrSlice_DeepCopy(t *testing.T) {
	original := make(nbrSlice, 2, 4)
	original[0] = NameBackReferences{Gvk: resid.FromKind("A"), Referrers: types.FsSlice{{Path: "a"}}}
	original[1] = NameBackReferences{Gvk: resid.FromKind("B"), Referrers: types.FsSlice{{Path: "b"}}}

	copied := original.DeepCopy()

	original, _ = original.mergeOne(NameBackReferences{Gvk: resid.FromKind("C"), Referrers: types.FsSlice{{Path: "c"}}})

	// perform mutations which should not affect original
	copied.Swap(0, 1)
	copied[0].Referrers[0].Path = "very b" // ensure Referrers are not shared
	_, _ = copied.mergeOne(NameBackReferences{Gvk: resid.FromKind("D"), Referrers: types.FsSlice{{Path: "d"}}})

	// if DeepCopy does not work, original would be {very b,a,d} instead of {a,b,c}
	expected := nbrSlice{
		{Gvk: resid.FromKind("A"), Referrers: types.FsSlice{{Path: "a"}}},
		{Gvk: resid.FromKind("B"), Referrers: types.FsSlice{{Path: "b"}}},
		{Gvk: resid.FromKind("C"), Referrers: types.FsSlice{{Path: "c"}}},
	}

	if !reflect.DeepEqual(original, expected) {
		t.Fatalf("original affected by mutations to copied object:\ngot\t%+v,\nexpected: %+v", original, expected)
	}
}
