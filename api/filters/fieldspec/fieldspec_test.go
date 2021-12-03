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

func TestFilter_Filter(t *testing.T) {
	testCases := map[string]struct {
		input     string
		expected  string
		filter    fieldspec.Filter
		fieldSpec string
		error     string
	}{
		"path not found": {
			fieldSpec: `
path: a/b
group: foo
kind: Bar
`,
			input: `
apiVersion: foo
kind: Bar
xxx:
`,
			expected: `
apiVersion: foo
kind: Bar
xxx:
`,
			filter: fieldspec.Filter{
				SetValue: filtersutil.SetScalar("e"),
			},
		},
		"empty path": {
			fieldSpec: `
group: foo
version: v1
kind: Bar
`,
			input: `
apiVersion: foo/v1
kind: Bar
xxx:
`,
			expected: `
apiVersion: foo
kind: Bar
xxx:
`,
			error: `considering field '' of object
apiVersion: foo/v1
kind: Bar
xxx:
metadata:
  annotations:
    internal.config.kubernetes.io/annotations-migration-resource-id: '0'
: cannot set or create an empty field name`,
			filter: fieldspec.Filter{
				SetValue: filtersutil.SetScalar("e"),
			},
		},

		"update": {
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

		"update-kind-not-match": {
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

		"update-group-not-match": {
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

		"update-version-not-match": {
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

		"bad-version": {
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

		"bad-meta": {
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
			expected: `
a:
  b: c
`,
			filter: fieldspec.Filter{
				SetValue: filtersutil.SetScalar("e"),
			},
		},

		"miss-match-type": {
			fieldSpec: `
path: a/b/c
kind: Bar
`,
			input: `
kind: Bar
a:
  b: a
`,
			error: `considering field 'a/b/c' of object
kind: Bar
a:
  b: a
metadata:
  annotations:
    internal.config.kubernetes.io/annotations-migration-resource-id: '0'
: expected sequence or mapping node`,
			filter: fieldspec.Filter{
				SetValue: filtersutil.SetScalar("e"),
			},
		},

		"add": {
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

		"update-in-sequence": {
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
		"empty-sequence-no-create": {
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
		"empty-sequence-create": {
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

		"group v1": {
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

		"version v1": {
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

		"successfully set field on array entry no sequence hint": {
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

		"successfully set field on array entry with sequence hint": {
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
		"failure to set field on array entry with sequence hint in path": {
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

		"failure to set field on array entry, no sequence hint in path": {
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
		"fieldname with slash '/'": {
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
		"fieldname with multiple '/'": {
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

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			err := yaml.Unmarshal([]byte(tc.fieldSpec), &tc.filter.FieldSpec)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			out := &bytes.Buffer{}
			rw := &kio.ByteReadWriter{
				Reader:                bytes.NewBufferString(tc.input),
				Writer:                out,
				OmitReaderAnnotations: true,
			}

			// run the filter
			err = kio.Pipeline{
				Inputs:  []kio.Reader{rw},
				Filters: []kio.Filter{kio.FilterAll(tc.filter)},
				Outputs: []kio.Writer{rw},
			}.Execute()
			if tc.error != "" {
				if !assert.EqualError(t, err, tc.error) {
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
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(out.String())) {
				t.FailNow()
			}
		})
	}
}
