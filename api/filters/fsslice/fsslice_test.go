// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fsslice_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	. "sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type TestCase struct {
	input    string
	expected string
	filter   Filter
	fsSlice  string
	error    string
}

var tests = map[string]TestCase{
	"empty": {
		fsSlice: `
`,
		input: `
apiVersion: foo/v1
kind: Bar
`,
		expected: `
apiVersion: foo/v1
kind: Bar
`,
		filter: Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},
	"two": {
		fsSlice: `
- path: a/b
  group: foo
  version: v1
  create: true
  kind: Bar
- path: q/r[]/s/t
  group: foo
  version: v1
  create: true
  kind: Bar
`,
		input: `
apiVersion: foo/v1
kind: Bar
q:
  r:
  - s: {}
`,
		expected: `
apiVersion: foo/v1
kind: Bar
q:
  r:
  - s: {t: e}
a:
  b: e
`,
		filter: Filter{
			SetValue:   filtersutil.SetScalar("e"),
			CreateKind: yaml.ScalarNode,
		},
	},
}

func TestFilter(t *testing.T) {
	for name := range tests {
		test := tests[name]
		t.Run(name, func(t *testing.T) {
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
