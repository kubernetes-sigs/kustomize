// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

var mergeTests = []struct {
	name     string
	original FsSlice
	incoming FsSlice
	err      error
	result   FsSlice
}{
	{
		"normal",
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
		FsSlice{
			{
				Path:               "home",
				Gvk:                resid.Gvk{Group: "beans"},
				CreateIfNotPresent: false,
			},
		},
		nil,
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "home",
				Gvk:                resid.Gvk{Group: "beans"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		"ignore copy",
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
		},
		nil,
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		"error on conflict",
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
		FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: true,
			},
		},
		fmt.Errorf("hey"),
		FsSlice{},
	},
}

func TestFsSlice_MergeAll(t *testing.T) {
	for _, item := range mergeTests {
		result, err := item.original.MergeAll(item.incoming)
		if item.err == nil {
			if err != nil {
				t.Fatalf("test %s: unexpected err %v", item.name, err)
			}
			if !reflect.DeepEqual(item.result, result) {
				t.Fatalf("test %s: expected: %v\n but got: %v\n",
					item.name, item.result, result)
			}
		} else {
			if err == nil {
				t.Fatalf("test %s: expected err: %v", item.name, err)
			}
			if !strings.Contains(err.Error(), "conflicting fieldspecs") {
				t.Fatalf("test %s: unexpected err: %v", item.name, err)
			}
		}
	}
}

func TestFsSlice_DeepCopy(t *testing.T) {
	original := make(FsSlice, 2, 4)
	original[0] = FieldSpec{Path: "a"}
	original[1] = FieldSpec{Path: "b"}

	copied := original.DeepCopy()

	original, _ = original.MergeOne(FieldSpec{Path: "c"})

	// perform mutations which should not affect original
	copied.Swap(0, 1)
	_, _ = copied.MergeOne(FieldSpec{Path: "d"})

	// if DeepCopy does not work, original would be {b,a,d} instead of {a,b,c}
	expected := FsSlice{
		{Path: "a"},
		{Path: "b"},
		{Path: "c"},
	}
	if !reflect.DeepEqual(original, expected) {
		t.Fatalf("original affected by mutations to copied object:\ngot\t%+v,\nexpected: %+v", original, expected)
	}
}
