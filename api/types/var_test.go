// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/resid"
)

func TestGVK(t *testing.T) {
	type testcase struct {
		data     string
		expected resid.Gvk
	}

	testcases := []testcase{
		{
			data: `
apiVersion: v1
kind: Secret
name: my-secret
`,
			expected: resid.Gvk{Group: "", Version: "v1", Kind: "Secret"},
		},
		{
			data: `
apiVersion: myapps/v1
kind: MyKind
name: my-kind
`,
			expected: resid.Gvk{Group: "myapps", Version: "v1", Kind: "MyKind"},
		},
		{
			data: `
version: v2
kind: MyKind
name: my-kind
`,
			expected: resid.Gvk{Version: "v2", Kind: "MyKind"},
		},
	}

	for _, tc := range testcases {
		var targ Target
		err := yaml.Unmarshal([]byte(tc.data), &targ)
		if err != nil {
			t.Fatalf("Unexpected error %v", err)
		}
		if !reflect.DeepEqual(targ.GVK(), tc.expected) {
			t.Fatalf("Expected %v, but got %v", tc.expected, targ.GVK())
		}
	}
}

func TestDefaulting(t *testing.T) {
	v := &Var{
		Name: "SOME_VARIABLE_NAME",
		ObjRef: Target{
			Gvk: resid.Gvk{
				Version: "v1",
				Kind:    "Secret",
			},
			Name: "my-secret",
		},
	}
	v.Defaulting()
	if v.FieldRef.FieldPath != defaultFieldPath {
		t.Fatalf("expected %s, got %v",
			defaultFieldPath, v.FieldRef.FieldPath)
	}
}

func TestVarSet(t *testing.T) {
	set := NewVarSet()
	vars := []Var{
		{
			Name: "SHELLVARS",
			ObjRef: Target{
				APIVersion: "v7",
				Gvk:        resid.Gvk{Kind: "ConfigMap"},
				Name:       "bash"},
		},
		{
			Name: "BACKEND",
			ObjRef: Target{
				APIVersion: "v7",
				Gvk:        resid.Gvk{Kind: "Deployment"},
				Name:       "myTiredBackend"},
		},
		{
			Name: "AWARD",
			ObjRef: Target{
				APIVersion: "v7",
				Gvk:        resid.Gvk{Kind: "Service"},
				Name:       "nobelPrize"},
			FieldRef: FieldSelector{FieldPath: "some.arbitrary.path"},
		},
	}
	err := set.MergeSlice(vars)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	for _, v := range vars {
		if !set.Contains(v) {
			t.Fatalf("set %v should contain var %v", set.AsSlice(), v)
		}
	}
	set2 := NewVarSet()
	err = set2.MergeSet(set)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = set2.MergeSlice(vars)
	if err == nil {
		t.Fatalf("expected err")
	}
	if !strings.Contains(err.Error(), "var 'SHELLVARS' already encountered") {
		t.Fatalf("unexpected err: %v", err)
	}
	v := set2.Get("BACKEND")
	if v == nil {
		t.Fatalf("expected var")
	}
	// Confirm defaulting.
	if v.FieldRef.FieldPath != defaultFieldPath {
		t.Fatalf("unexpected field path: %v", v.FieldRef.FieldPath)
	}
	// Confirm sorting.
	names := set2.AsSlice()
	if names[0].Name != "AWARD" ||
		names[1].Name != "BACKEND" ||
		names[2].Name != "SHELLVARS" {
		t.Fatalf("unexpected order in : %v", names)
	}
}

func TestVarSetCopy(t *testing.T) {
	set1 := NewVarSet()
	vars := []Var{
		{Name: "First"},
		{Name: "Second"},
		{Name: "Third"},
	}
	err := set1.MergeSlice(vars)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// Confirm copying
	set2 := set1.Copy()
	for _, varInSet1 := range set1.AsSlice() {
		if v := set2.Get(varInSet1.Name); v == nil {
			t.Fatalf("set %v should contain a Var named %s", set2.AsSlice(), varInSet1)
		} else if !set2.Contains(*v) {
			t.Fatalf("set %v should contain %v", set2.AsSlice(), v)
		}
	}
	// Confirm that the copy is deep
	w := Var{Name: "Only in set2"}
	set2.Merge(w)
	if !set2.Contains(w) {
		t.Fatalf("set %v should contain %v", set2.AsSlice(), w)
	}
	if set1.Contains(w) {
		t.Fatalf("set %v should not contain %v", set1.AsSlice(), w)
	}
}
