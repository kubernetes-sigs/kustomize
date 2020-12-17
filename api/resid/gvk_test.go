// Copyright 2018 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var equalsTests = []struct {
	x1 Gvk
	x2 Gvk
}{
	{Gvk{Group: "a", Version: "b", Kind: "c"},
		Gvk{Group: "a", Version: "b", Kind: "c"}},
	{Gvk{Version: "b", Kind: "c"},
		Gvk{Version: "b", Kind: "c"}},
	{Gvk{Kind: "c"},
		Gvk{Kind: "c"}},
}

func TestEquals(t *testing.T) {
	for _, hey := range equalsTests {
		if !hey.x1.Equals(hey.x2) {
			t.Fatalf("%v should equal %v", hey.x1, hey.x2)
		}
	}
}

var lessThanTests = []struct {
	x1 Gvk
	x2 Gvk
}{
	{Gvk{Group: "a", Version: "b", Kind: "CustomResourceDefinition"},
		Gvk{Group: "a", Version: "b", Kind: "RoleBinding"}},
	{Gvk{Group: "a", Version: "b", Kind: "Namespace"},
		Gvk{Group: "a", Version: "b", Kind: "ClusterRole"}},
	{Gvk{Group: "a", Version: "b", Kind: "a"},
		Gvk{Group: "a", Version: "b", Kind: "b"}},
	{Gvk{Group: "a", Version: "b", Kind: "Namespace"},
		Gvk{Group: "a", Version: "c", Kind: "Namespace"}},
	{Gvk{Group: "a", Version: "c", Kind: "Namespace"},
		Gvk{Group: "b", Version: "c", Kind: "Namespace"}},
	{Gvk{Group: "b", Version: "c", Kind: "Namespace"},
		Gvk{Group: "a", Version: "c", Kind: "ClusterRole"}},
	{Gvk{Group: "a", Version: "c", Kind: "Namespace"},
		Gvk{Group: "a", Version: "b", Kind: "ClusterRole"}},
	{Gvk{Group: "a", Version: "d", Kind: "Namespace"},
		Gvk{Group: "b", Version: "c", Kind: "Namespace"}},
	{Gvk{Group: "a", Version: "b", Kind: orderFirst[len(orderFirst)-1]},
		Gvk{Group: "a", Version: "b", Kind: orderLast[0]}},
	{Gvk{Group: "a", Version: "b", Kind: orderFirst[len(orderFirst)-1]},
		Gvk{Group: "a", Version: "b", Kind: "CustomKindX"}},
	{Gvk{Group: "a", Version: "b", Kind: "CustomKindX"},
		Gvk{Group: "a", Version: "b", Kind: orderLast[0]}},
	{Gvk{Group: "a", Version: "b", Kind: "CustomKindA"},
		Gvk{Group: "a", Version: "b", Kind: "CustomKindB"}},
	{Gvk{Group: "a", Version: "b", Kind: "CustomKindX"},
		Gvk{Group: "a", Version: "b", Kind: "MutatingWebhookConfiguration"}},
	{Gvk{Group: "a", Version: "b", Kind: "MutatingWebhookConfiguration"},
		Gvk{Group: "a", Version: "b", Kind: "ValidatingWebhookConfiguration"}},
	{Gvk{Group: "a", Version: "b", Kind: "CustomKindX"},
		Gvk{Group: "a", Version: "b", Kind: "ValidatingWebhookConfiguration"}},
	{Gvk{Group: "a", Version: "b", Kind: "APIService"},
		Gvk{Group: "a", Version: "b", Kind: "ValidatingWebhookConfiguration"}},
	{Gvk{Group: "a", Version: "b", Kind: "Service"},
		Gvk{Group: "a", Version: "b", Kind: "APIService"}},
	{Gvk{Group: "a", Version: "b", Kind: "Endpoints"},
		Gvk{Group: "a", Version: "b", Kind: "Service"}},
}

func TestIsLessThan1(t *testing.T) {
	for _, hey := range lessThanTests {
		if !hey.x1.IsLessThan(hey.x2) {
			t.Fatalf("%v should be less than %v", hey.x1, hey.x2)
		}
		if hey.x2.IsLessThan(hey.x1) {
			t.Fatalf("%v should not be less than %v", hey.x2, hey.x1)
		}
	}
}

var stringTests = []struct {
	x Gvk
	s string
	r string
}{
	{Gvk{}, "~G_~V_~K", ""},
	{Gvk{Kind: "k"}, "~G_~V_k", "k"},
	{Gvk{Version: "v"}, "~G_v_~K", "v"},
	{Gvk{Version: "v", Kind: "k"}, "~G_v_k", "v_k"},
	{Gvk{Group: "g"}, "g_~V_~K", "g"},
	{Gvk{Group: "g", Kind: "k"}, "g_~V_k", "g_k"},
	{Gvk{Group: "g", Version: "v"}, "g_v_~K", "g_v"},
	{Gvk{Group: "g", Version: "v", Kind: "k"}, "g_v_k", "g_v_k"},
}

func TestString(t *testing.T) {
	for _, hey := range stringTests {
		assert.Equal(t, hey.s, hey.x.String())
	}
}

func TestApiVersion(t *testing.T) {
	for _, hey := range []struct {
		x   Gvk
		exp string
	}{
		{Gvk{}, ""},
		{Gvk{Kind: "k"}, ""},
		{Gvk{Version: "v"}, "v"},
		{Gvk{Version: "v", Kind: "k"}, "v"},
		{Gvk{Group: "g"}, "g/"},
		{Gvk{Group: "g", Kind: "k"}, "g/"},
		{Gvk{Group: "g", Version: "v"}, "g/v"},
		{Gvk{Group: "g", Version: "v", Kind: "k"}, "g/v"},
	} {
		assert.Equal(t, hey.exp, hey.x.ApiVersion())
	}
}

func TestStringWoEmptyField(t *testing.T) {
	for _, hey := range stringTests {
		assert.Equal(t, hey.r, hey.x.StringWoEmptyField())
	}
}

func TestParseGroupVersion(t *testing.T) {
	tests := []struct {
		input string
		g     string
		v     string
	}{
		{input: "", g: "", v: ""},
		{input: "v1", g: "", v: "v1"},
		{input: "apps/v1", g: "apps", v: "v1"},
		{input: "/v1", g: "", v: "v1"},
		{input: "apps/", g: "apps", v: ""},
		{input: "/apps/", g: "", v: "apps/"},
	}
	for _, tc := range tests {
		g, v := ParseGroupVersion(tc.input)
		assert.Equal(t, tc.g, g, tc.input)
		assert.Equal(t, tc.v, v, tc.input)
	}
}

func TestSelectByGVK(t *testing.T) {
	type testCase struct {
		description string
		in          Gvk
		filter      *Gvk
		expected    bool
	}
	testCases := []testCase{
		{
			description: "nil filter",
			in:          Gvk{},
			filter:      nil,
			expected:    true,
		},
		{
			description: "gvk matches",
			in: Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			expected: true,
		},
		{
			description: "group doesn't matches",
			in: Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &Gvk{
				Group:   "group2",
				Version: "version1",
				Kind:    "kind1",
			},
			expected: false,
		},
		{
			description: "version doesn't matches",
			in: Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &Gvk{
				Group:   "group1",
				Version: "version2",
				Kind:    "kind1",
			},
			expected: false,
		},
		{
			description: "kind doesn't matches",
			in: Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind2",
			},
			expected: false,
		},
		{
			description: "no version in filter",
			in: Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &Gvk{
				Group:   "group1",
				Version: "",
				Kind:    "kind1",
			},
			expected: true,
		},
		{
			description: "only kind is set in filter",
			in: Gvk{
				Group:   "group1",
				Version: "version1",
				Kind:    "kind1",
			},
			filter: &Gvk{
				Group:   "",
				Version: "",
				Kind:    "kind1",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		filtered := tc.in.IsSelected(tc.filter)
		assert.Equal(t, tc.expected, filtered, tc.description)
	}
}

func TestIsNamespaceableKind(t *testing.T) {
	testCases := []struct {
		name     string
		gvk      Gvk
		expected bool
	}{
		{
			"namespaceable resource",
			Gvk{Group: "apps", Version: "v1", Kind: "Deployment"},
			true,
		},
		{
			"clusterscoped resource",
			Gvk{Group: "", Version: "v1", Kind: "Namespace"},
			false,
		},
		{
			"unknown resource (should default to namespaceable)",
			Gvk{Group: "example1.com", Version: "v1", Kind: "Bar"},
			true,
		},
		{
			"unknown resource (should default to namespaceable)",
			Gvk{Group: "apps", Version: "v1", Kind: "ClusterRoleBinding"},
			true,
		},
	}

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			isNamespaceable := test.gvk.IsNamespaceableKind()
			assert.Equal(t, test.expected, isNamespaceable)
		})
	}
}
