// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
package fieldspec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestParseGV(t *testing.T) {
	testCases := map[string]struct {
		input           string
		expectedGroup   string
		expectedVersion string
	}{
		"empty": {
			input:           "",
			expectedGroup:   "",
			expectedVersion: "",
		},
		"certSigning": {
			input:           "certificates.k8s.io/v1beta1",
			expectedGroup:   "certificates.k8s.io",
			expectedVersion: "v1beta1",
		},
		"extensions": {
			input:           "extensions/v1beta1",
			expectedGroup:   "extensions",
			expectedVersion: "v1beta1",
		},
		"normal": {
			input:           "apps/v1",
			expectedGroup:   "apps",
			expectedVersion: "v1",
		},
		"justApps": {
			input:           "apps",
			expectedGroup:   "apps",
			expectedVersion: "",
		},
		"coreV1": {
			input:           "v1",
			expectedGroup:   "",
			expectedVersion: "v1",
		},
		"coreV2": {
			input:           "v2",
			expectedGroup:   "",
			expectedVersion: "v2",
		},
		"coreV2Beta1": {
			input:           "v2beta1",
			expectedGroup:   "",
			expectedVersion: "v2beta1",
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			group, version := parseGV(tc.input)
			if !assert.Equal(t, tc.expectedGroup, group) {
				t.FailNow()
			}
			if !assert.Equal(t, tc.expectedVersion, version) {
				t.FailNow()
			}
		})
	}
}

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
			expected: resid.Gvk{Group: "", Version: "v1", Kind: ""},
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
