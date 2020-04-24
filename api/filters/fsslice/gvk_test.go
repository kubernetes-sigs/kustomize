// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
package fsslice

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestGetGVK(t *testing.T) {
	testCases := map[string]struct {
		input      string
		expected   resid.Gvk
		parseError string
		metaError  string
	}{
		"empty": {
			input: `
`,
			parseError: "EOF",
		},
		"junk": {
			input: `
congress: effective
`,
			metaError: "missing Resource metadata",
		},
		"normal": {
			input: `
apiVersion: apps/v1
kind: Deployment
`,
			expected: resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"},
		},
		"apiVersionOnlyWithSlash": {
			input: `
apiVersion: apps/v1
`,
			expected: resid.Gvk{Group: "apps", Version: "v1", Kind: ""},
		},
		// When apiVersion is just "v1" (not, say, "apps/v1"), that
		// could be interpreted as Group="", Version="v1"
		// (implying the original "core" api group) or the other way around
		// (Group="v1", Version="").
		// At the time of writing, fsslice.go does the latter -
		// might have to change that.
		"apiVersionOnlyNoSlash1": {
			input: `
apiVersion: apps
`,
			expected: resid.Gvk{Group: "apps", Version: "", Kind: ""},
		},
		"apiVersionOnlyNoSlash2": {
			input: `
apiVersion: v1
`,
			expected: resid.Gvk{Group: "v1", Version: "", Kind: ""},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			obj, err := yaml.Parse(tc.input)
			if len(tc.parseError) != 0 {
				if err == nil {
					t.Error("expected parse error")
					return
				}
				if !strings.Contains(err.Error(), tc.parseError) {
					t.Errorf("expected parse err '%s', got '%v'", tc.parseError, err)
				}
				return
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			meta, err := obj.GetMeta()
			if len(tc.metaError) != 0 {
				if err == nil {
					t.Error("expected meta error")
					return
				}
				if !strings.Contains(err.Error(), tc.metaError) {
					t.Errorf("expected meta err '%s', got '%v'", tc.metaError, err)
				}
				return
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			gvk := GetGVK(meta)
			if !assert.Equal(t, tc.expected, gvk) {
				t.FailNow()
			}
		})
	}
}
