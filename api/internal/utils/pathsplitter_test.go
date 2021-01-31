// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/internal/utils"
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
		assert.Equal(t, tc.exp, PathSplitter(tc.path))
	}
}
