// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fsslice_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
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

	{
		name: "group v1",
		fsSlice: `
- path: a/b
  group: v1
  create: true
  kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
`,
		expected: `
apiVersion: v1
kind: Bar
`,
		filter: fsslice.Filter{
			SetValue:   fsslice.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	{
		name: "version v1",
		fsSlice: `
- path: a/b
  version: v1
  create: true
  kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
`,
		expected: `
apiVersion: v1
kind: Bar
a:
  b: e
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
