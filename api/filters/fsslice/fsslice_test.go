// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fsslice_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type TestCase struct {
	name     string
	input    string
	expected string
	filter   fsslice.Filter
	fsSlice  string
	error    string
}

var tests = []TestCase{
	{
		name: "update",
		fsSlice: `
- path: a/b
  group: foo
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta1
kind: Bar
a:
  b: c
`,
		expected: `
apiVersion: foo/v1beta1
kind: Bar
a:
  b: e
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	{
		name: "update-kind-not-match",
		fsSlice: `
- path: a/b
  group: foo
  kind: Bar1
`,
		input: `
apiVersion: foo/v1beta1
kind: Bar2
a:
  b: c
`,
		expected: `
apiVersion: foo/v1beta1
kind: Bar2
a:
  b: c
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	{
		name: "update-group-not-match",
		fsSlice: `
- path: a/b
  group: foo1
  kind: Bar
`,
		input: `
apiVersion: foo2/v1beta1
kind: Bar
a:
  b: c
`,
		expected: `
apiVersion: foo2/v1beta1
kind: Bar
a:
  b: c
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	{
		name: "update-version-not-match",
		fsSlice: `
- path: a/b
  group: foo
  version: v1beta1
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta2
kind: Bar
a:
  b: c
`,
		expected: `
apiVersion: foo/v1beta2
kind: Bar
a:
  b: c
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	{
		name: "bad-version",
		fsSlice: `
- path: a/b
  group: foo
  version: v1beta1
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta2/something
kind: Bar
a:
  b: c
`,
		expected: `
apiVersion: foo/v1beta2/something
kind: Bar
a:
  b: c
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	{
		name: "bad-meta",
		fsSlice: `
- path: a/b
  group: foo
  version: v1beta1
  kind: Bar
`,
		input: `
a:
  b: c
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
		error: "missing Resource metadata",
	},

	{
		name: "miss-match-type",
		fsSlice: `
- path: a/b/c
  kind: Bar
`,
		input: `
kind: Bar
a:
  b: a
`,
		error: "obj kind: Bar\na:\n  b: a\n at path a/b/c: unsupported yaml node",
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	{
		name: "add",
		fsSlice: `
- path: a/b/c/d
  group: foo
  create: true
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta1
kind: Bar
a: {}
`,
		expected: `
apiVersion: foo/v1beta1
kind: Bar
a: {b: {c: {d: e}}}
`,
		filter: fsslice.Filter{
			SetValue:   fsslice.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	{
		name: "update-in-sequence",
		fsSlice: `
- path: a/b[]/c/d
  group: foo
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta1
kind: Bar
a:
  b:
  - c:
      d: a
`,
		expected: `
apiVersion: foo/v1beta1
kind: Bar
a:
  b:
  - c:
      d: e
`,
		filter: fsslice.Filter{
			SetValue: fsslice.SetScalar("e"),
		},
	},

	// Don't create a sequence
	{
		name: "empty-sequence-no-create",
		fsSlice: `
- path: a/b[]/c/d
  group: foo
  create: true
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta1
kind: Bar
a: {}
`,
		expected: `
apiVersion: foo/v1beta1
kind: Bar
a: {}
`,
		filter: fsslice.Filter{
			SetValue:   fsslice.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	// Create a new field for an element in a sequence
	{
		name: "empty-sequence-create",
		fsSlice: `
- path: a/b[]/c/d
  group: foo
  create: true
  kind: Bar
`,
		input: `
apiVersion: foo/v1beta1
kind: Bar
a:
  b:
  - c: {}
`,
		expected: `
apiVersion: foo/v1beta1
kind: Bar
a:
  b:
  - c: {d: e}
`,
		filter: fsslice.Filter{
			SetValue:   fsslice.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},
}

func TestFilter_Filter(t *testing.T) {
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			err := yaml.Unmarshal([]byte(test.fsSlice), &test.filter.FsSlice)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			out := &bytes.Buffer{}
			rw := &kio.ByteReadWriter{
				Reader:                bytes.NewBufferString(test.input),
				Writer:                out,
				OmitReaderAnnotations: true,
			}

			// run the filter
			err = kio.Pipeline{
				Inputs:  []kio.Reader{rw},
				Filters: []kio.Filter{kio.FilterAll(test.filter)},
				Outputs: []kio.Writer{rw},
			}.Execute()
			if test.error != "" {
				if !assert.EqualError(t, err, test.error) {
					t.FailNow()
				}
				// stop rest of test
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// check results
			if !assert.Equal(t,
				strings.TrimSpace(test.expected),
				strings.TrimSpace(out.String())) {
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
			gvk := fsslice.GetGVK(meta)
			if !assert.Equal(t, tc.expected, gvk) {
				t.FailNow()
			}
		})
	}
}
