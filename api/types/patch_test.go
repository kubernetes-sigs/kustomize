// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func TestPatchEquals(t *testing.T) {
	selector := Selector{
		ResId: resid.ResId{
			Gvk: resid.Gvk{
				Group:   "group",
				Version: "version",
				Kind:    "kind",
			},
			Name:      "name",
			Namespace: "namespace",
		},
		LabelSelector:      "selector",
		AnnotationSelector: "selector",
	}
	type testcase struct {
		patch1 Patch
		patch2 Patch
		expect bool
		name   string
	}
	testcases := []testcase{
		{
			name:   "empty patches",
			patch1: Patch{},
			patch2: Patch{},
			expect: true,
		},
		{
			name: "full patches",
			patch1: Patch{
				Path:  "foo",
				Patch: "bar",
				Target: &Selector{
					ResId: resid.ResId{
						Gvk: resid.Gvk{
							Group:   "group",
							Version: "version",
							Kind:    "kind",
						},
						Name:      "name",
						Namespace: "namespace",
					},
					LabelSelector:      "selector",
					AnnotationSelector: "selector",
				},
			},
			patch2: Patch{
				Path:  "foo",
				Patch: "bar",
				Target: &Selector{
					ResId: resid.ResId{
						Gvk: resid.Gvk{
							Group:   "group",
							Version: "version",
							Kind:    "kind",
						},
						Name:      "name",
						Namespace: "namespace",
					},
					LabelSelector:      "selector",
					AnnotationSelector: "selector",
				},
			},
			expect: true,
		},
		{
			name: "same target",
			patch1: Patch{
				Path:   "foo",
				Patch:  "bar",
				Target: &selector,
			},
			patch2: Patch{
				Path:   "foo",
				Patch:  "bar",
				Target: &selector,
			},
			expect: true,
		},
		{
			name: "omit target",
			patch1: Patch{
				Path:  "foo",
				Patch: "bar",
			},
			patch2: Patch{
				Path:  "foo",
				Patch: "bar",
			},
			expect: true,
		},
		{
			name: "one nil target",
			patch1: Patch{
				Path:   "foo",
				Patch:  "bar",
				Target: &selector,
			},
			patch2: Patch{
				Path:  "foo",
				Patch: "bar",
			},
			expect: false,
		},
		{
			name: "different path",
			patch1: Patch{
				Path: "foo",
			},
			patch2: Patch{
				Path: "bar",
			},
			expect: false,
		},
	}

	for _, tc := range testcases {
		if tc.expect != tc.patch1.Equals(tc.patch2) {
			t.Fatalf("%s: unexpected result %v", tc.name, !tc.expect)
		}
	}
}
