// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type TestCase struct {
	name      string
	input     string
	expected  string
	filter    fieldspec.Filter
	fieldSpec string
	error     string
}

var tests = []TestCase{
	{
		name: "update",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	{
		name: "update-kind-not-match",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	{
		name: "update-group-not-match",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	{
		name: "update-version-not-match",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	{
		name: "bad-version",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	{
		name: "bad-meta",
		fieldSpec: `
path: a/b
group: foo
version: v1beta1
kind: Bar
`,
		input: `
a:
  b: c
`,
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
		error: "missing Resource metadata",
	},

	{
		name: "miss-match-type",
		fieldSpec: `
path: a/b/c
kind: Bar
`,
		input: `
kind: Bar
a:
  b: a
`,
		error: "obj 'kind: Bar\na:\n  b: a\n' at path 'a/b/c': " +
			"expected sequence or mapping node",
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	{
		name: "add",
		fieldSpec: `
path: a/b/c/d
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
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	{
		name: "update-in-sequence",
		fieldSpec: `
path: a/b[]/c/d
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
		filter: fieldspec.Filter{
			SetValue: filtersutil.SetScalar("e"),
		},
	},

	// Don't create a sequence
	{
		name: "empty-sequence-no-create",
		fieldSpec: `
path: a/b[]/c/d
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
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	// Create a new field for an element in a sequence
	{
		name: "empty-sequence-create",
		fieldSpec: `
path: a/b[]/c/d
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
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	{
		name: "group v1",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},

	{
		name: "version v1",
		fieldSpec: `
path: a/b
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
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},
	{
		name: "successfully set field on array entry no sequence hint",
		fieldSpec: `
path: spec/containers/image
version: v1
kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
spec:
  containers:
  - image: foo
`,
		expected: `
apiVersion: v1
kind: Bar
spec:
  containers:
  - image: bar
`,
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("bar"),
			CreateKind: yaml.ScalarNode,
		},
	},
	{
		name: "successfully set field on array entry with sequence hint",
		fieldSpec: `
path: spec/containers[]/image
version: v1
kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
spec:
  containers:
  - image: foo
`,
		expected: `
apiVersion: v1
kind: Bar
spec:
  containers:
  - image: bar
`,
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("bar"),
			CreateKind: yaml.ScalarNode,
		},
	},
	{
		name: "failure to set field on array entry with sequence hint in path",
		fieldSpec: `
path: spec/containers[]/image
version: v1
kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
spec:
  containers:
`,
		expected: `
apiVersion: v1
kind: Bar
spec:
  containers: []
`,
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("bar"),
			CreateKind: yaml.ScalarNode,
		},
	},
	{
		name: "failure to set field on array entry, no sequence hint in path",
		fieldSpec: `
path: spec/containers/image
version: v1
kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
spec:
  containers:
`,
		expected: `
apiVersion: v1
kind: Bar
spec:
  containers:
`,
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("bar"),
			CreateKind: yaml.ScalarNode,
		},
	},
	{
		name: "filedname with slash '/'",
		fieldSpec: `
path: a/b\/c/d
version: v1
kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
a:
  b/c:
    d: foo
`,
		expected: `
apiVersion: v1
kind: Bar
a:
  b/c:
    d: bar
`,
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("bar"),
			CreateKind: yaml.ScalarNode,
		},
	},
	{
		name: "filedname with multiple '/'",
		fieldSpec: `
path: a/b\/c/d\/e/f
version: v1
kind: Bar
`,
		input: `
apiVersion: v1
kind: Bar
a:
  b/c:
    d/e:
      f: foo
`,
		expected: `
apiVersion: v1
kind: Bar
a:
  b/c:
    d/e:
      f: bar
`,
		filter: fieldspec.Filter{
			SetValue:   filtersutil.SetScalar("bar"),
			CreateKind: yaml.ScalarNode,
		},
	},
}

func TestFilter_Filter(t *testing.T) {
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			err := yaml.Unmarshal([]byte(test.fieldSpec), &test.filter.FieldSpec)
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
