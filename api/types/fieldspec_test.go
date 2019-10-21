// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/resid"
	. "sigs.k8s.io/kustomize/api/types"
)

func TestPathSlice(t *testing.T) {
	type path struct {
		input  string
		parsed []string
	}
	paths := []path{
		{
			input:  "spec/metadata/annotations",
			parsed: []string{"spec", "metadata", "annotations"},
		},
		{
			input:  `metadata/annotations/nginx.ingress.kubernetes.io\/auth-secret`,
			parsed: []string{"metadata", "annotations", "nginx.ingress.kubernetes.io/auth-secret"},
		},
	}
	for _, p := range paths {
		fs := FieldSpec{Path: p.input}
		actual := fs.PathSlice()
		if !reflect.DeepEqual(actual, p.parsed) {
			t.Fatalf("expected %v, but got %v", p.parsed, actual)
		}
	}
}

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
