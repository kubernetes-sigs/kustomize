// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/utils"
)

func TestPathSplitter(t *testing.T) {
	for _, tc := range []struct {
		exp  []string
		path string
	}{
		{
			path: "",
			exp:  []string{""},
		},
		{
			path: "s",
			exp:  []string{"s"},
		},
		{
			path: "a/b/c",
			exp:  []string{"a", "b", "c"},
		},
		{
			path: `a/b[]/c`,
			exp:  []string{"a", "b[]", "c"},
		},
		{
			path: `a/b\/c/d\/e/f`,
			exp:  []string{"a", "b/c", "d/e", "f"},
		},
		{
			// The actual reason for this.
			path: `metadata/annotations/nginx.ingress.kubernetes.io\/auth-secret`,
			exp: []string{
				"metadata",
				"annotations",
				"nginx.ingress.kubernetes.io/auth-secret"},
		},
	} {
		assert.Equal(t, tc.exp, PathSplitter(tc.path, "/"))
	}
}

func TestSmarterPathSplitter(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected []string
	}{
		"simple": {
			input:    "spec.replicas",
			expected: []string{"spec", "replicas"},
		},
		"sequence": {
			input:    "spec.data.[name=first].key",
			expected: []string{"spec", "data", "[name=first]", "key"},
		},
		"key, value with . prefix": {
			input:    "spec.data.[.name=.first].key",
			expected: []string{"spec", "data", "[.name=.first]", "key"},
		},
		"key, value with . suffix": {
			input:    "spec.data.[name.=first.].key",
			expected: []string{"spec", "data", "[name.=first.]", "key"},
		},
		"multiple '.' in value": {
			input:    "spec.data.[name=f.i.r.s.t.].key",
			expected: []string{"spec", "data", "[name=f.i.r.s.t.]", "key"},
		},
		"with escaped delimiter": {
			input:    `spec\.replicas`,
			expected: []string{`spec.replicas`},
		},
		"unmatched bracket": {
			input:    "spec.data.[name=f.i.[r.s.t..key",
			expected: []string{"spec", "data", "[name=f.i.[r.s.t..key"},
		},
		"mapping value with .": {
			input:    "metadata.annotations.[a.b.c/d.e.f-g.]",
			expected: []string{"metadata", "annotations", "a.b.c/d.e.f-g."},
		},
	}
	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expected, SmarterPathSplitter(tc.input, "."))
		})
	}
}
